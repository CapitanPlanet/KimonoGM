using Microsoft.JSInterop;
using JanuszSimulator.Models;
using System.Net.Http.Json;

namespace JanuszSimulator.Services;

public class GameEngine
{
    private readonly HttpClient _http;
    private readonly IJSRuntime _js;
    public GameState GameState { get; set; } = new();
    private CancellationTokenSource? _reactionCts;
    private bool _musicStarted = false;

    private AvatarSystem _avatarSystem = new();
    private readonly string[] _allowedPrefixes = { "av_", "bg_", "re_" };
    private string _currentDayId = "day1";
    private ProjectManifest? _projectManifest;
    private List<string> _allDayIds = new() { "day1" };
    private HashSet<string> _loadedDays = new();

    public event Action? StateChanged;

    public GameEngine(HttpClient http, IJSRuntime js)
    {
        _http = http;
        _js = js;
    }

    public class GameDataWrapper { public List<Scene> scenes { get; set; } = new(); }

    public class AvatarSystem
    {
        public string Default { get; set; } = "";
        public string? FallbackDefault { get; set; }
        public List<AvatarRule> Rules { get; set; } = new();
    }
    public class AvatarRule
    {
        public string Id { get; set; } = "";
        public string Name { get; set; } = "";
        public string Use { get; set; } = "";
        public string? Fallback { get; set; }
        public Dictionary<string, Dictionary<string, int>>? If { get; set; }
        public int Priority { get; set; } = 0;
    }
    public class ProjectManifest
    {
        public string gameName { get; set; } = "";
        public string startDay { get; set; } = "day1";
        public string startScene { get; set; } = "start";
        public List<string> days { get; set; } = new();
        public AvatarSystem avatarSystem { get; set; } = new();
        public StatsSystem statsSystem { get; set; } = new();
    }
    public class StatsSystem { public List<StatDef> stats { get; set; } = new(); }

    private string NormalizeStatId(string raw)
    {
        if (string.IsNullOrWhiteSpace(raw)) return raw;
        var up = raw.Trim().ToUpperInvariant();
        return up switch
        {
            "A" => "CEBULA", "B" => "WSTYD", "C" => "PORTFEL", "D" => "REPUTACJA", "E" => "REPUTACJA",
            "CEBULA" => "CEBULA", "WSTYD" => "WSTYD", "PORTFEL" => "PORTFEL", "REPUTACJA" => "REPUTACJA",
            _ => up
        };
    }

    private bool IsAllowedAsset(string path)
    {
        if (string.IsNullOrWhiteSpace(path)) return false;
        var file = Path.GetFileName(path).ToLowerInvariant().Trim();
        bool hasPrefix = _allowedPrefixes.Any(p => file.StartsWith(p));
        bool isAllowedExt = file.EndsWith(".jpg") || file.EndsWith(".jpeg") || file.EndsWith(".png") || file.EndsWith(".webp");
        return hasPrefix && isAllowedExt;
    }

    private string NormalizeBg(string? raw)
    {
        if (string.IsNullOrWhiteSpace(raw)) return "images/bg_front.jpg";
        raw = raw.Trim().Replace("\\", "/").Trim();
        var fileName = Path.GetFileName(raw);
        if (string.IsNullOrWhiteSpace(fileName)) return "images/bg_front.jpg";
        return $"images/{fileName}";
    }

    public async Task LoadGame()
    {
        GameState.AllScenes = new List<Scene>();
        _loadedDays.Clear();
        _currentDayId = "day1";
        ProjectManifest? project = null;

        // 1. LOAD project.janproj
        try
        {
            var respProj = await _http.GetAsync("data/project.janproj");
            if (respProj.IsSuccessStatusCode)
            {
                project = await respProj.Content.ReadFromJsonAsync<ProjectManifest>();
                Console.WriteLine($"[Game] project.janproj: {project?.statsSystem?.stats?.Count} statow");
            }
            else
            {
                var resp2 = await _http.GetAsync("project.janproj");
                if (resp2.IsSuccessStatusCode) project = await resp2.Content.ReadFromJsonAsync<ProjectManifest>();
            }
        }
        catch (Exception ex) { Console.WriteLine($"[Game] brak project.janproj: {ex.Message}"); }

        // manifest z edytora
        try
        {
            var man = await _http.GetFromJsonAsync<ManifestFile>("data/_manifest.json");
            if (man != null && man.days.Any())
            {
                _allDayIds = man.days;
                Console.WriteLine($"[Game] manifest.json: days=[{string.Join(",", _allDayIds)}] start={man.startDay}/{man.startScene}");
                if (project == null) project = new ProjectManifest();
                project.startDay = man.startDay ?? project.startDay;
                project.startScene = man.startScene ?? project.startScene;
                project.days = man.days;
            }
        }
        catch (Exception ex) { Console.WriteLine($"[Game] brak manifest.json: {ex.Message}"); }

        _projectManifest = project;
        if (project?.days != null && project.days.Any()) _allDayIds = project.days;

        // 2. STATY
        if (project?.statsSystem?.stats != null && project.statsSystem.stats.Any())
        {
            GameState.StatDefs = project.statsSystem.stats.Select(s => new StatDef
            {
                id = NormalizeStatId(s.id),
                name = string.IsNullOrWhiteSpace(s.name) ? s.id : s.name,
                initial = s.initial
            }).ToList();
            Console.WriteLine($"[Game] Staty z projektu: {string.Join(",", GameState.StatDefs.Select(s => s.name))}");
        }

        // 3. AVATAR
        if (project?.avatarSystem != null)
        {
            var def = project.avatarSystem.Default;
            var filteredRules = project.avatarSystem.Rules.Where(r => IsAllowedAsset(r.Use)).ToList();
            _avatarSystem = new AvatarSystem { Default = IsAllowedAsset(def) ? def : def ?? "", Rules = filteredRules };
            Console.WriteLine($"[Avatar] z project.janproj: default={_avatarSystem.Default} regul={filteredRules.Count}");
        }
        else
        {
            _avatarSystem = new AvatarSystem { Default = "", Rules = new() };
        }

        _currentDayId = project?.startDay ?? "day1";
        var startSceneId = project?.startScene ?? "start";

        try
        {
            Console.WriteLine($"[Game] Probuje ladowac data/{_currentDayId}.json");
            var day1 = await _http.GetFromJsonAsync<List<Scene>>($"data/{_currentDayId}.json");
            if (day1 != null && day1.Any())
            {
                GameState.AllScenes.AddRange(day1);
                _loadedDays.Add(_currentDayId);
                Console.WriteLine($"[Game] Zaladowano {day1.Count} scen z {_currentDayId}.json");
            }
        }
        catch (Exception ex)
        {
            Console.WriteLine($"[Game] Nie udalo sie zaladowac {_currentDayId}.json: {ex.Message}");
            try
            {
                var data = await _http.GetFromJsonAsync<GameDataWrapper>("data/scenes.json") ?? new();
                GameState.AllScenes = data.scenes;
            }
            catch { }
        }

        GameState.Stats = new DynamicStats();
        GameState.Flags = new();
        GameState.ReactionImage = "";
        GameState.ReactionText = "";

        if (!GameState.StatDefs.Any())
        {
            GameState.StatDefs = new List<StatDef>
            {
                new() { id = "CEBULA", name = "Cebula", initial = 50 },
                new() { id = "WSTYD", name = "Wstyd", initial = 0 },
            };
        }
        foreach (var def in GameState.StatDefs) GameState.Stats.Set(def.id, def.initial);

        if (GameState.AllScenes.Any())
        {
            var start = GameState.AllScenes.FirstOrDefault(s => s.Id == startSceneId) ?? GameState.AllScenes.First();
            Console.WriteLine($"[Game] Start sceny: {start.Id} z {GameState.AllScenes.Count} scen");
            LoadScene(start.Id);
        }
    }

    public class ManifestFile
    {
        public List<string> days { get; set; } = new();
        public string startDay { get; set; } = "day1";
        public string startScene { get; set; } = "start";
        public string version { get; set; } = "";
    }

    public async Task LoadDay(string dayId)
    {
        if (string.IsNullOrWhiteSpace(dayId) || dayId == "END") return;
        if (_loadedDays.Contains(dayId))
        {
            _currentDayId = dayId;
            Console.WriteLine($"[Game] Dzień {dayId} już załadowany, przełączam");
            return;
        }
        try
        {
            Console.WriteLine($"[Game] Ładuję dzień {dayId}.json");
            var scenes = await _http.GetFromJsonAsync<List<Scene>>($"data/{dayId}.json");
            if (scenes != null && scenes.Any())
            {
                _currentDayId = dayId;
                // DODAJEMY a nie nadpisujemy - żeby wsteczne skoki działały
                foreach (var s in scenes)
                {
                    if (!GameState.AllScenes.Any(x => x.Id == s.Id)) GameState.AllScenes.Add(s);
                    else
                    {
                        var idx = GameState.AllScenes.FindIndex(x => x.Id == s.Id);
                        GameState.AllScenes[idx] = s;
                    }
                }
                _loadedDays.Add(dayId);
                Console.WriteLine($"[Game] Załadowano {scenes.Count} scen z {dayId}. Razem {GameState.AllScenes.Count}");
                var start = scenes.FirstOrDefault(s => s.Id.StartsWith("start"))?.Id ?? scenes.First().Id;
                LoadScene(start);
            }
        }
        catch (Exception ex) { Console.WriteLine($"[Game] Nie udało się załadować {dayId}: {ex.Message}"); }
    }

    // NOWE: znajdź dzień który zawiera scenę
    private string? FindDayForScene(string sceneId)
    {
        // jeśli już załadowana - zwróć null (nie trzeba ładować)
        if (GameState.AllScenes.Any(s => s.Id == sceneId)) return null;
        // przeszukaj manifest - wszystkie dni jeszcze niezaładowane
        // ale nie wiemy który dzień ma tę scenę bez ładowania, więc próbujemy wszystkie
        return null;
    }

    private async Task<bool> EnsureSceneLoaded(string sceneId)
    {
        if (GameState.AllScenes.Any(s => s.Id == sceneId)) return true;
        Console.WriteLine($"[Game] Scena {sceneId} nie w pamięci, próbuję załadować pozostałe dni: {string.Join(",", _allDayIds.Where(d => !_loadedDays.Contains(d)))}");
        foreach (var dayId in _allDayIds.Where(d => !_loadedDays.Contains(d)))
        {
            try
            {
                var scenes = await _http.GetFromJsonAsync<List<Scene>>($"data/{dayId}.json");
                if (scenes == null) continue;
                foreach (var s in scenes)
                {
                    if (!GameState.AllScenes.Any(x => x.Id == s.Id)) GameState.AllScenes.Add(s);
                }
                _loadedDays.Add(dayId);
                Console.WriteLine($"[Game] Doładowałem {dayId}: {scenes.Count} scen, razem {GameState.AllScenes.Count}");
                if (GameState.AllScenes.Any(s => s.Id == sceneId)) return true;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[Game] Nie udało się doładować {dayId}: {ex.Message}");
            }
        }
        return GameState.AllScenes.Any(s => s.Id == sceneId);
    }

    public void LoadScene(string sceneId)
    {
        if (sceneId == "END_DAY") return;
        if (sceneId == "END")
        {
            GameState.IsFinished = true;
            Console.WriteLine("[Game] END - koniec gry");
            NotifyStateChanged();
            return;
        }
        GameState.CurrentScene = GameState.AllScenes.FirstOrDefault(s => s.Id == sceneId);
        if (GameState.CurrentScene == null) { Console.WriteLine($"[Game] Nie znaleziono sceny {sceneId} w {GameState.AllScenes.Count} załadowanych. Dni załadowane: {string.Join(",", _loadedDays)}"); return; }
        GameState.BackgroundImage = NormalizeBg(GameState.CurrentScene.Background);
        GameState.SceneTitle = GameState.CurrentScene.SceneTitle ?? "";
        GameState.NarrationText = GameState.CurrentScene.Text ?? "";
        GameState.CurrentChoices = GameState.CurrentScene.Choices ?? new List<Choice>();
        GameState.ReactionText = ""; GameState.ReactionImage = "";
        GameState.IsFinished = GameState.CurrentChoices.Count == 0;
        UpdateAvatar(); NotifyStateChanged();
    }

    public bool CanSelectChoice(Choice choice)
    {
        if (choice.FlagsRequired.Any() && !choice.FlagsRequired.All(f => GameState.Flags.Contains(f))) return false;
        foreach (var kv in choice.MinStats) if (GameState.Stats.Get(NormalizeStatId(kv.Key)) < kv.Value) return false;
        foreach (var kv in choice.MaxStats) if (GameState.Stats.Get(NormalizeStatId(kv.Key)) > kv.Value) return false;
        if (choice.KosztPortfel.HasValue && GameState.Stats.Get("PORTFEL") < choice.KosztPortfel.Value) return false;
        return true;
    }

    public IEnumerable<Choice> GetAvailableChoices()
        => GameState.CurrentChoices?.Where(c => CanSelectChoice(c)) ?? Enumerable.Empty<Choice>();

    public async Task MakeChoice(Choice choice)
    {
        if (!CanSelectChoice(choice))
        {
            GameState.ReactionText = string.IsNullOrWhiteSpace(choice.FailText) ? "Nie możesz tego zrobić, Janusz." : choice.FailText;
            GameState.ReactionImage = ""; NotifyStateChanged();
            await Task.Delay(5000); GameState.ReactionText = ""; GameState.ReactionImage = ""; NotifyStateChanged(); return;
        }
        if (!_musicStarted) { _musicStarted = true; try { await _js.InvokeVoidAsync("JanuszAudio.startMusic", "sounds/s1.mp3"); } catch { } }
        _reactionCts?.Cancel(); ApplyChoiceEffects(choice);
        GameState.ReactionText = string.IsNullOrWhiteSpace(choice.ReactionText) ? "" : choice.ReactionText;
        var reactionImgRaw = choice.ReactionImage;
        if (string.IsNullOrWhiteSpace(reactionImgRaw)) reactionImgRaw = "";
        else if (!IsAllowedAsset(reactionImgRaw)) reactionImgRaw = "";
        GameState.ReactionImage = reactionImgRaw; NotifyStateChanged();
        if (!string.IsNullOrWhiteSpace(choice.SoundFile)) _ = _js.InvokeVoidAsync("JanuszAudio.playVoice", $"sounds/{choice.SoundFile}");
        if (!string.IsNullOrWhiteSpace(GameState.ReactionText) || !string.IsNullOrWhiteSpace(GameState.ReactionImage))
        {
            _reactionCts = new CancellationTokenSource();
            try { await Task.Delay(5000, _reactionCts.Token); } catch (TaskCanceledException) { } finally { HideReaction(); }
        }
        if (!string.IsNullOrWhiteSpace(choice.Next))
        {
            // FIX 1: END_DAY stary system
            if (choice.Next == "END_DAY")
            {
                var endScene = GameState.CurrentScene;
                var nextDay = endScene?.NextDayId ?? endScene?.NextDay ?? "";
                if (string.IsNullOrWhiteSpace(nextDay) || nextDay == "END") { Console.WriteLine("[Game] END_DAY bez NextDay - koniec gry"); GameState.IsFinished = true; NotifyStateChanged(); return; }
                Console.WriteLine($"[Game] END_DAY -> ładuję {nextDay}"); await LoadDay(nextDay); return;
            }

            // FIX 2: END - koniec gry
            if (choice.Next == "END")
            {
                Console.WriteLine("[Game] END - koniec gry"); GameState.IsFinished = true; NotifyStateChanged(); return;
            }

            // FIX 3: jeśli to koniec dnia (IsEndDay) i ma NextDayId, załaduj dzień przed przejściem
            if (GameState.CurrentScene?.IsEndDay == true || GameState.CurrentScene?.Type == "end_of_day")
            {
                var nextDay = GameState.CurrentScene?.NextDayId ?? GameState.CurrentScene?.NextDay;
                if (!string.IsNullOrWhiteSpace(nextDay) && nextDay != "END" && !_loadedDays.Contains(nextDay))
                {
                    Console.WriteLine($"[Game] Koniec dnia {_currentDayId} -> ładuję {nextDay} przed {choice.Next}");
                    await LoadDay(nextDay);
                    // LoadDay już zrobi LoadScene(start), ale jeśli choice.Next jest inny niż start, nadpiszemy niżej
                }
            }

            // FIX 4: jeśli scena docelowa nie jest załadowana, spróbuj załadować wszystkie pozostałe dni
            if (!GameState.AllScenes.Any(s => s.Id == choice.Next))
            {
                var loaded = await EnsureSceneLoaded(choice.Next);
                if (!loaded)
                {
                    Console.WriteLine($"[Game] Nadal nie znaleziono sceny {choice.Next} po doładowaniu wszystkich dni");
                    return;
                }
            }

            LoadScene(choice.Next);
        }
    }

    private void ApplyChoiceEffects(Choice choice)
    {
        foreach (var kv in choice.Stats) { var key = NormalizeStatId(kv.Key); GameState.Stats.Set(key, GameState.Stats.Get(key) + kv.Value); }
        if (choice.Cebula != 0) GameState.Stats.Set("CEBULA", GameState.Stats.Get("CEBULA") + choice.Cebula);
        if (choice.Wstyd != 0) GameState.Stats.Set("WSTYD", GameState.Stats.Get("WSTYD") + choice.Wstyd);
        if (choice.Portfel != 0) GameState.Stats.Set("PORTFEL", GameState.Stats.Get("PORTFEL") + choice.Portfel);
        if (choice.Reputacja != 0) GameState.Stats.Set("REPUTACJA", GameState.Stats.Get("REPUTACJA") + choice.Reputacja);
        if (choice.KosztPortfel.HasValue) GameState.Stats.Set("PORTFEL", GameState.Stats.Get("PORTFEL") - choice.KosztPortfel.Value);
        foreach (var flag in choice.FlagsSet) GameState.Flags.Add(flag);
        foreach (var key in GameState.Stats.Values.Keys.ToList())
        {
            var v = GameState.Stats.Values[key];
            if (key == "CEBULA" || key == "WSTYD" || key == "REPUTACJA") GameState.Stats.Values[key] = Math.Clamp(v, 0, 100);
            else if (key == "PORTFEL") GameState.Stats.Values[key] = Math.Max(0, v);
        }
    }

    public void SkipReaction() => _reactionCts?.Cancel();
    public void HideReaction() { GameState.ReactionImage = ""; GameState.ReactionText = ""; NotifyStateChanged(); }

    private void UpdateAvatar()
    {
        foreach (var rule in _avatarSystem.Rules.OrderByDescending(r => r.Priority))
        {
            if (rule.If == null || rule.If.Count == 0) continue;
            bool matches = true;
            foreach (var kv in rule.If)
            {
                var currentValue = GameState.Stats.Get(NormalizeStatId(kv.Key));
                foreach (var cond in kv.Value)
                {
                    if (cond.Key == "gte" && currentValue < cond.Value) matches = false;
                    if (cond.Key == "lte" && currentValue > cond.Value) matches = false;
                    if (cond.Key == "gt" && currentValue <= cond.Value) matches = false;
                    if (cond.Key == "lt" && currentValue >= cond.Value) matches = false;
                    if (cond.Key == "eq" && currentValue != cond.Value) matches = false;
                }
            }
            if (matches && IsAllowedAsset(rule.Use)) { GameState.JanuszImage = rule.Use; return; }
        }
        GameState.JanuszImage = _avatarSystem.Default ?? "";
    }

    private void NotifyStateChanged() => StateChanged?.Invoke();
}
