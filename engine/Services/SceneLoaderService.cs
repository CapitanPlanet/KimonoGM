using System.Net.Http.Json;
using MakerEngine.Models;
namespace MakerEngine.Services;

public class SceneLoaderService
{
    private readonly HttpClient _http;
    public ProjectManifest? Manifest { get; private set; }
    public List<string> AllDayIds { get; private set; } = new() { "day1" };
    public HashSet<string> LoadedDays { get; } = new();
    public string CurrentDayId { get; private set; } = "day1";

    public SceneLoaderService(HttpClient http) => _http = http;

    public async Task<(List<Scene> scenes, string startDay, string startScene)> LoadInitialAsync()
    {
        ProjectManifest? project = null;
        try
        {
            var resp = await _http.GetAsync("data/project.janproj");
            if (resp.IsSuccessStatusCode) project = await resp.Content.ReadFromJsonAsync<ProjectManifest>();
            else { var r2 = await _http.GetAsync("project.janproj"); if (r2.IsSuccessStatusCode) project = await r2.Content.ReadFromJsonAsync<ProjectManifest>(); }
        } catch {}

        try
        {
            var man = await _http.GetFromJsonAsync<ManifestFile>("data/_manifest.json");
            if (man!= null && man.days.Any())
            {
                AllDayIds = man.days;
                if (project == null) project = new ProjectManifest();
                project.startDay = man.startDay?? project.startDay;
                project.startScene = man.startScene?? project.startScene;
                project.days = man.days;
            }
        } catch {}

        Manifest = project;
        if (project?.days!= null && project.days.Any()) AllDayIds = project.days;

        CurrentDayId = project?.startDay?? "day1";
        var startScene = project?.startScene?? "start";
        var all = new List<Scene>();

        try
        {
            var day1 = await _http.GetFromJsonAsync<List<Scene>>($"data/{CurrentDayId}.json");
            if (day1!= null && day1.Any()) { all.AddRange(day1); LoadedDays.Add(CurrentDayId); }
        }
        catch
        {
            try { var data = await _http.GetFromJsonAsync<GameDataWrapper>("data/scenes.json"); if (data!= null) all = data.scenes; } catch {}
        }
        return (all, CurrentDayId, startScene);
    }

    public async Task<List<Scene>> LoadDayAsync(string dayId)
    {
        var scenes = await _http.GetFromJsonAsync<List<Scene>>($"data/{dayId}.json");
        if (scenes!= null) { CurrentDayId = dayId; LoadedDays.Clear(); LoadedDays.Add(dayId); }
        return scenes?? new();
    }

    public async Task<bool> EnsureSceneLoadedAsync(List<Scene> allScenes, string sceneId)
    {
        if (allScenes.Any(s => s.Id == sceneId)) return true;
        foreach (var dayId in AllDayIds.Where(d =>!LoadedDays.Contains(d)))
        {
            try
            {
                var scenes = await _http.GetFromJsonAsync<List<Scene>>($"data/{dayId}.json");
                if (scenes!= null && scenes.Any())
                {
                    allScenes.AddRange(scenes.Where(s =>!allScenes.Any(e => e.Id == s.Id)));
                    LoadedDays.Add(dayId);
                    if (scenes.Any(s => s.Id == sceneId)) return true;
                }
            } catch {}
        }
        return allScenes.Any(s => s.Id == sceneId);
    }
}
