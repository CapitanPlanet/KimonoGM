using Microsoft.JSInterop;
using MakerEngine.Models;

namespace MakerEngine.Services;

public class GameEngine
{
    private readonly HttpClient _http;
    private readonly IJSRuntime _js;
    private readonly StatsService _stats = new();
    private readonly AvatarService _avatar = new();
    private readonly SceneLoaderService _loader;

    public GameState GameState { get; } = new();
    private CancellationTokenSource? _reactionCts;
    private bool _musicStarted = false;
    private AvatarSystem _avatarSystem = new();

    public event Action? StateChanged;

    public GameEngine(HttpClient http, IJSRuntime js)
    {
        _http = http;
        _js = js;
        _loader = new SceneLoaderService(http);
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
        var (scenes, startDay, startScene) = await _loader.LoadInitialAsync();
        GameState.AllScenes = scenes;
        var manifest = _loader.Manifest;

        if (manifest?.statsSystem?.stats!= null && manifest.statsSystem.stats.Any())
        {
            GameState.StatDefs = manifest.statsSystem.stats.Select(s => new StatDef
            {
                id = _stats.NormalizeStatId(s.id),
                name = string.IsNullOrWhiteSpace(s.name)? s.id : s.name,
                initial = s.initial
            }).ToList();
        }

        if (manifest?.avatarSystem!= null)
        {
            var def = manifest.avatarSystem.Default;
            var filtered = _avatar.FilterRules(manifest.avatarSystem.Rules);
            _avatarSystem = new AvatarSystem { Default = _avatar.IsAllowedAsset(def)? def : def?? "", Rules = filtered };
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
            var start = GameState.AllScenes.FirstOrDefault(s => s.Id == startScene)?? GameState.AllScenes.First();
            LoadScene(start.Id);
        }
    }

    public void LoadScene(string sceneId)
    {
        if (sceneId == "END_DAY") return;
        GameState.CurrentScene = GameState.AllScenes.FirstOrDefault(s => s.Id == sceneId);
        if (GameState.CurrentScene == null) return;
        GameState.BackgroundImage = NormalizeBg(GameState.CurrentScene.Background);
        GameState.SceneTitle = GameState.CurrentScene.SceneTitle?? "";
        GameState.NarrationText = GameState.CurrentScene.Text?? "";
        GameState.CurrentChoices = GameState.CurrentScene.Choices?? new List<Choice>();
        GameState.ReactionText = ""; GameState.ReactionImage = "";
        GameState.IsFinished = GameState.CurrentChoices.Count == 0;
        _avatar.Update(GameState, _avatarSystem, NormalizeBg);
        StateChanged?.Invoke();
    }

    public bool CanSelectChoice(Choice c) => _stats.CanSelect(GameState, c);
    public IEnumerable<Choice> GetAvailableChoices() => GameState.CurrentChoices?.Where(c => CanSelectChoice(c))?? Enumerable.Empty<Choice>();

    public async Task MakeChoice(Choice choice)
    {
        if (!CanSelectChoice(choice))
        {
            GameState.ReactionText = string.IsNullOrWhiteSpace(choice.FailText)? "Nie możesz tego zrobić, Janusz." : choice.FailText;
            GameState.ReactionImage = ""; StateChanged?.Invoke();
            await Task.Delay(5000); GameState.ReactionText = ""; GameState.ReactionImage = ""; StateChanged?.Invoke(); return;
        }

        if (!_musicStarted) _musicStarted = true;
        _reactionCts?.Cancel();
        _stats.ApplyEffects(GameState, choice);
        GameState.ReactionText = choice.ReactionText?? "";
        GameState.ReactionImage = _avatar.IsAllowedAsset(choice.ReactionImage?? "")? choice.ReactionImage?? "" : "";
        StateChanged?.Invoke();

        if (!string.IsNullOrWhiteSpace(choice.SoundFile))
        {
            var lower = choice.SoundFile.ToLowerInvariant();
            var blacklist = new[] { "s1.mp3", "s2.mp3", "s3.mp3", "old_music.mp3", "muzyka.mp3" };
            if (!blacklist.Contains(lower)) _ = _js.InvokeVoidAsync("JanuszAudio.playVoice", $"sounds/{choice.SoundFile}");
        }

        if (!string.IsNullOrWhiteSpace(GameState.ReactionText) ||!string.IsNullOrWhiteSpace(GameState.ReactionImage))
        {
            _reactionCts = new CancellationTokenSource();
            try { await Task.Delay(5000, _reactionCts.Token); } catch (TaskCanceledException) {} finally { HideReaction(); }
        }

        if (string.IsNullOrWhiteSpace(choice.Next)) return;
        if (choice.Next == "END_DAY" || GameState.CurrentScene?.IsEndDay == true)
        {
            var nextDay = GameState.CurrentScene?.NextDayId?? GameState.CurrentScene?.NextDay?? "";
            if (string.IsNullOrWhiteSpace(nextDay) || nextDay == "END") { GameState.IsFinished = true; StateChanged?.Invoke(); return; }
            var scenes = await _loader.LoadDayAsync(nextDay);
            GameState.AllScenes = scenes;
            LoadScene(scenes.FirstOrDefault(s => s.Id.StartsWith("start"))?.Id?? scenes.First().Id);
            return;
        }
        if (choice.Next == "END") { GameState.IsFinished = true; StateChanged?.Invoke(); return; }

        if (!await _loader.EnsureSceneLoadedAsync(GameState.AllScenes, choice.Next)) return;
        LoadScene(choice.Next);
    }

    public void SkipReaction() => _reactionCts?.Cancel();
    public void HideReaction() { GameState.ReactionImage = ""; GameState.ReactionText = ""; StateChanged?.Invoke(); }
}
