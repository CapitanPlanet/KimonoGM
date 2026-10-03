using MakerEngine.Models;
namespace MakerEngine.Services;

public class AvatarService
{
    private readonly string[] _allowedPrefixes = { "av_", "bg_", "re_" };
    private bool IsAllowed(string path)
    {
        if (string.IsNullOrWhiteSpace(path)) return false;
        var file = Path.GetFileName(path).ToLowerInvariant().Trim();
        return _allowedPrefixes.Any(p => file.StartsWith(p)) && (file.EndsWith(".jpg") || file.EndsWith(".jpeg") || file.EndsWith(".png") || file.EndsWith(".webp"));
    }

    public void Update(GameState gs, AvatarSystem system, Func<string,string> normalize)
    {
        foreach (var rule in system.Rules.OrderByDescending(r => r.Priority))
        {
            if (rule.If == null || rule.If.Count == 0) continue;
            bool matches = true;
            foreach (var kv in rule.If)
            {
                var cur = gs.Stats.Get(kv.Key.ToUpperInvariant());
                foreach (var cond in kv.Value)
                {
                    if (cond.Key == "gte" && cur < cond.Value) matches = false;
                    if (cond.Key == "lte" && cur > cond.Value) matches = false;
                    if (cond.Key == "gt" && cur <= cond.Value) matches = false;
                    if (cond.Key == "lt" && cur >= cond.Value) matches = false;
                    if (cond.Key == "eq" && cur!= cond.Value) matches = false;
                }
            }
            if (matches && IsAllowed(rule.Use)) { gs.JanuszImage = rule.Use; return; }
        }
        gs.JanuszImage = system.Default?? "";
    }

    public List<AvatarRule> FilterRules(List<AvatarRule> rules) => rules.Where(r => IsAllowed(r.Use)).ToList();
    public bool IsAllowedAsset(string path) => IsAllowed(path);
}
