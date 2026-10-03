using MakerEngine.Models;
namespace MakerEngine.Services;

public class StatsService
{
    public string NormalizeStatId(string raw)
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

    public bool CanSelect(GameState gs, Choice choice)
    {
        if (choice.FlagsRequired.Any() &&!choice.FlagsRequired.All(f => gs.Flags.Contains(f))) return false;
        foreach (var kv in choice.MinStats) if (gs.Stats.Get(NormalizeStatId(kv.Key)) < kv.Value) return false;
        foreach (var kv in choice.MaxStats) if (gs.Stats.Get(NormalizeStatId(kv.Key)) > kv.Value) return false;
        if (choice.KosztPortfel.HasValue && gs.Stats.Get("PORTFEL") < choice.KosztPortfel.Value) return false;
        return true;
    }

    public void ApplyEffects(GameState gs, Choice choice)
    {
        foreach (var kv in choice.Stats) { var key = NormalizeStatId(kv.Key); gs.Stats.Set(key, gs.Stats.Get(key) + kv.Value); }
        if (choice.Cebula!= 0) gs.Stats.Set("CEBULA", gs.Stats.Get("CEBULA") + choice.Cebula);
        if (choice.Wstyd!= 0) gs.Stats.Set("WSTYD", gs.Stats.Get("WSTYD") + choice.Wstyd);
        if (choice.Portfel!= 0) gs.Stats.Set("PORTFEL", gs.Stats.Get("PORTFEL") + choice.Portfel);
        if (choice.Reputacja!= 0) gs.Stats.Set("REPUTACJA", gs.Stats.Get("REPUTACJA") + choice.Reputacja);
        if (choice.KosztPortfel.HasValue) gs.Stats.Set("PORTFEL", gs.Stats.Get("PORTFEL") - choice.KosztPortfel.Value);
        foreach (var flag in choice.FlagsSet) gs.Flags.Add(flag);
        foreach (var key in gs.Stats.Values.Keys.ToList())
        {
            var v = gs.Stats.Values[key];
            if (key == "CEBULA" || key == "WSTYD" || key == "REPUTACJA") gs.Stats.Values[key] = Math.Clamp(v, 0, 100);
            else if (key == "PORTFEL") gs.Stats.Values[key] = Math.Max(0, v);
        }
    }
}
