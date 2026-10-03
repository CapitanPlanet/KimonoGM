namespace MakerEngine.Services;

public class AssetWhitelistService
{
    private readonly string[] _prefixes = { "av_", "bg_", "re_" };
    private readonly string[] _ext = { ".jpg", ".jpeg", ".png", ".webp" };
    public AssetWhitelistService() { }
    public AssetWhitelistService(string[] prefixes) => _prefixes = prefixes;
    public bool IsAllowed(string path)
    {
        if (string.IsNullOrWhiteSpace(path)) return false;
        var file = Path.GetFileName(path).ToLowerInvariant().Trim();
        return _prefixes.Any(p => file.StartsWith(p)) && _ext.Any(e => file.EndsWith(e));
    }
}