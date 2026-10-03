namespace MakerEngine.Models;

public class ProjectManifest
{
    public string gameName { get; set; } = "";
    public string startDay { get; set; } = "day1";
    public string startScene { get; set; } = "start";
    public List<string> days { get; set; } = new();
    public AvatarSystem avatarSystem { get; set; } = new();
    public StatsSystem statsSystem { get; set; } = new();
}
public class ManifestFile
{
    public List<string> days { get; set; } = new();
    public string? startDay { get; set; }
    public string? startScene { get; set; }
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
