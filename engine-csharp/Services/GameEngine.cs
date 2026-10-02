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
        _currentDayId = "day1";
        ProjectManifest? project = null;

        // 1. LOAD project.janproj - TWOJE staty QQQQ, WWW itp
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
                // fallback - szukaj w root
                var resp2 = await _http.GetAsync("project.janproj");
                if (resp2.IsSuccessStatusCode) project = await resp2.Content.ReadFromJsonAsync<ProjectManifest>();
            }
        }
        catch (Exception ex) { Console.WriteLine($"[Game] brak project.janproj: {ex.Message}"); }

        // 2. STATY Z PROJEKTU - bez hardcoda
        if (project?.statsSystem?.stats!= null && project.statsSystem.stats.Any())
        {
            GameState.StatDefs = project.statsSystem.stats.Select(s => new StatDef
            {
                id = NormalizeStatId(s.id), // klucz normalizowany do lookupu
                name = string.IsNullOrWhiteSpace(s.name)? s.id : s.name, // nazwa oryginalna QQQQ
                initial = s.initial
            }).ToList();
            Console.WriteLine($"[Game] Staty z projektu: {string.Join(",", GameState.StatDefs.Select(s=>s.name))}");
        }

        // 3. AVATAR Z PROJEKTU
        if (project?.avatarSystem!= null)
        {
            var def = project.avatarSystem.Default;
            var filteredRules = project.avatarSystem.Rules.Where(r => IsAllowedAsset(r.Use)).ToList();
            _avatarSystem = new AvatarSystem { Default = IsAllowedAsset(def)? def : def?? "", Rules = filteredRules };
            Console.WriteLine($"[Avatar] z project.janproj: default={_avatarSystem.Default} regul={filteredRules.Count}");
        }
        else
        {
            _avatarSystem = new AvatarSystem { Default = "", Rules = new() };
        }

        _currentDayId = project?.startDay?? "day1";
        var startSceneId = project?.startScene?? "start";

        try
        {
            Console.WriteLine($"[Game] Probuje ladowac data/{_currentDayId}.json");
            var day1 = await _http.GetFromJsonAsync<List<Scene>>($"data/{_currentDayId}.json");
            if (day1!= null && day1.Any())
            {
                GameState.AllScenes.AddRange(day1);
                Console.WriteLine($"[Game] Zaladowano {day1.Count} scen z {_currentDayId}.json");
            }
        }
        catch (Exception ex)
        {
            Console.WriteLine($"[Game] Nie udalo sie zaladowac {_currentDayId}.json: {ex.Message}");
            try
            {
                var data = await _http.GetFromJsonAsync<GameDataWrapper>("data/scenes.json")?? new();
                GameState.AllScenes = data.scenes;
            }
            catch {}
        }

        GameState.Stats = new DynamicStats();
        GameState.Flags = new();
        GameState.ReactionImage = "";
        GameState.ReactionText = "";

        if (!GameState.StatDefs.Any())
        {
            GameState.StatDefs = new List<StatDef>
            {
                new() { id = "CEBULA", name = "CEBULA", initial = 10 },
                new() { id = "WSTYD", name = "WSTYD", initial = 10 },
                new() { id = "PORTFEL", name = "PORTFEL", initial = 100 },
                new() { id = "REPUTACJA", name = "REPUTACJA", initial = 50 },
            };
        }
        foreach (var d in GameState.StatDefs) GameState.Stats.Values[NormalizeStatId(d.id)] = d.initial;

        var targetId = startSceneId;
        if (!GameState.AllScenes.Any(s => s.Id == targetId))
            targetId = GameState.AllScenes.FirstOrDefault(s => s.Id == targetId)?.Id
                   ?? GameState.AllScenes.FirstOrDefault(s => s.Id.StartsWith("start"))?.Id
                   ?? GameState.AllScenes.FirstOrDefault()?.Id?? "start";

        Console.WriteLine($"[Game] Start sceny: {targetId} z {GameState.AllScenes.Count} scen");
        LoadScene(targetId);
    }

    public async Task LoadDay(string dayId)
    {
        if (string.IsNullOrWhiteSpace(dayId)) dayId = "day1";
        dayId = dayId.Replace(".json", "").Trim();
        Console.WriteLine($"[Game] Ładuję dzień: {dayId}");
        try
        {
            var scenes = await _http.GetFromJsonAsync<List<Scene>>($"data/{dayId}.json");
            if (scenes!= null && scenes.Any())
            {
                _currentDayId = dayId;
                GameState.AllScenes = scenes;
                var start = scenes.FirstOrDefault(s => s.Id.StartsWith("start"))?.Id?? scenes.First().Id;
                LoadScene(start);
            }
        }
        catch (Exception ex) { Console.WriteLine($"[Game] Nie udało się załadować {dayId}: {ex.Message}"); }
    }

    public void LoadScene(string sceneId)
    {
        if (sceneId == "END_DAY") return;
        GameState.CurrentScene = GameState.AllScenes.FirstOrDefault(s => s.Id == sceneId);
        if (GameState.CurrentScene == null) { Console.WriteLine($"[Game] Nie znaleziono sceny {sceneId}"); return; }
        GameState.BackgroundImage = NormalizeBg(GameState.CurrentScene.Background);
        GameState.SceneTitle = GameState.CurrentScene.SceneTitle?? "";
        GameState.NarrationText = GameState.CurrentScene.Text?? "";
        GameState.CurrentChoices = GameState.CurrentScene.Choices?? new List<Choice>();
        GameState.ReactionText = ""; GameState.ReactionImage = "";
        GameState.IsFinished = GameState.CurrentChoices.Count == 0;
        UpdateAvatar(); NotifyStateChanged();
    }

    public bool CanSelectChoice(Choice choice)
    {
        if (choice.FlagsRequired.Any() &&!choice.FlagsRequired.All(f => GameState.Flags.Contains(f))) return false;
        foreach (var kv in choice.MinStats) if (GameState.Stats.Get(NormalizeStatId(kv.Key)) < kv.Value) return false;
        foreach (var kv in choice.MaxStats) if (GameState.Stats.Get(NormalizeStatId(kv.Key)) > kv.Value) return false;
        if (choice.KosztPortfel.HasValue && GameState.Stats.Get("PORTFEL") < choice.KosztPortfel.Value) return false;
        return true;
    }

    public IEnumerable<Choice> GetAvailableChoices()
        => GameState.CurrentChoices?.Where(c => CanSelectChoice(c))?? Enumerable.Empty<Choice>();

    public async Task MakeChoice(Choice choice)
    {
        if (!CanSelectChoice(choice))
        {
            GameState.ReactionText = string.IsNullOrWhiteSpace(choice.FailText)? "Nie możesz tego zrobić, Janusz." : choice.FailText;
            GameState.ReactionImage = ""; NotifyStateChanged();
            await Task.Delay(5000); GameState.ReactionText = ""; GameState.ReactionImage = ""; NotifyStateChanged(); return;
        }
        if (!_musicStarted) { _musicStarted = true; try { await _js.InvokeVoidAsync("JanuszAudio.startMusic", "sounds/s1.mp3"); } catch {} }
        _reactionCts?.Cancel(); ApplyChoiceEffects(choice);
        GameState.ReactionText = string.IsNullOrWhiteSpace(choice.ReactionText)? "" : choice.ReactionText;
        var reactionImgRaw = choice.ReactionImage;
        if (string.IsNullOrWhiteSpace(reactionImgRaw)) reactionImgRaw = "";
        else if (!IsAllowedAsset(reactionImgRaw)) reactionImgRaw = "";
        GameState.ReactionImage = reactionImgRaw; NotifyStateChanged();
        if (!string.IsNullOrWhiteSpace(choice.SoundFile)) _ = _js.InvokeVoidAsync("JanuszAudio.playVoice", $"sounds/{choice.SoundFile}");
        if (!string.IsNullOrWhiteSpace(GameState.ReactionText) ||!string.IsNullOrWhiteSpace(GameState.ReactionImage))
        {
            _reactionCts = new CancellationTokenSource();
            try { await Task.Delay(5000, _reactionCts.Token); } catch (TaskCanceledException) { } finally { HideReaction(); }
        }
        if (!string.IsNullOrWhiteSpace(choice.Next))
        {
            if (choice.Next == "END_DAY")
            {
                var endScene = GameState.CurrentScene;
                var nextDay = endScene?.NextDayId?? endScene?.NextDay?? "";
                if (string.IsNullOrWhiteSpace(nextDay)) { Console.WriteLine("[Game] END_DAY bez NextDay - koniec gry"); GameState.IsFinished = true; NotifyStateChanged(); return; }
                Console.WriteLine($"[Game] END_DAY -> ładuję {nextDay}"); await LoadDay(nextDay); return;
            }
            LoadScene(choice.Next);
        }
    }

    private void ApplyChoiceEffects(Choice choice)
    {
        foreach (var kv in choice.Stats) { var key = NormalizeStatId(kv.Key); GameState.Stats.Set(key, GameState.Stats.Get(key) + kv.Value); }
        if (choice.Cebula!= 0) GameState.Stats.Set("CEBULA", GameState.Stats.Get("CEBULA") + choice.Cebula);
        if (choice.Wstyd!= 0) GameState.Stats.Set("WSTYD", GameState.Stats.Get("WSTYD") + choice.Wstyd);
        if (choice.Portfel!= 0) GameState.Stats.Set("PORTFEL", GameState.Stats.Get("PORTFEL") + choice.Portfel);
        if (choice.Reputacja!= 0) GameState.Stats.Set("REPUTACJA", GameState.Stats.Get("REPUTACJA") + choice.Reputacja);
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
                    if (cond.Key == "eq" && currentValue!= cond.Value) matches = false;
                }
            }
            if (matches && IsAllowedAsset(rule.Use)) { GameState.JanuszImage = rule.Use; return; }
        }
        GameState.JanuszImage = _avatarSystem.Default?? "";
    }

    private void NotifyStateChanged() => StateChanged?.Invoke();
}