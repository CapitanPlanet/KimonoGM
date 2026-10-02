using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace JanuszTools;

public class BundleScenes
{
    public static int Main(string[] args)
    {
        try
        {
            var projectDir = args.Length > 0? args[0] : Directory.GetCurrentDirectory();
            Console.WriteLine($"[BUNDLER] ProjectDir: {projectDir}");

            var assetsDir = Path.Combine(projectDir, "Assets", "Data");
            var outDir = Path.Combine(projectDir, "wwwroot", "data");

            if (!Directory.Exists(assetsDir))
            {
                Console.WriteLine($"[BUNDLER ERROR] Brak folderu: {assetsDir}");
                return 1;
            }

            Directory.CreateDirectory(outDir);

            // 1. KOPIUJ project.janproj - SZUKAJ WSZĘDZIE
            var projectCandidates = new[] {
                Path.Combine(projectDir, "..", "games", "10005", "project.janproj"),
                Path.Combine(projectDir, "..", "..", "games", "10005", "project.janproj"),
                Path.Combine(projectDir, "Assets", "project.janproj"),
                Path.Combine(projectDir, "Assets", "Data", "..", "project.janproj"),
                Path.Combine(projectDir, "project.janproj"),
                Path.Combine(Directory.GetCurrentDirectory(), "..", "games", "10005", "project.janproj"),
                Path.Combine(Directory.GetCurrentDirectory(), "games", "10005", "project.janproj")
            };

            bool projFound = false;
            foreach (var src in projectCandidates)
            {
                var full = Path.GetFullPath(src);
                if (File.Exists(full))
                {
                    File.Copy(full, Path.Combine(outDir, "project.janproj"), true);
                    Console.WriteLine($"[BUNDLER] project.janproj copied from {full}");
                    projFound = true;
                    break;
                }
            }
            if (!projFound)
            {
                Console.WriteLine("[BUNDLER WARN] Nie znaleziono project.janproj - staty beda fallbackiem");
                // stworz pusty zeby nie bylo 404
                var dummy = @"{""gameName"":""10005"",""startDay"":""day1"",""startScene"":""start"",""avatarSystem"":{""default"":"""",""rules"":[]},""statsSystem"":{""stats"":[{""id"":""QQQQ"",""name"":""QQQQ"",""initial"":0},{""id"":""WWW"",""name"":""WWW"",""initial"":0},{""id"":""EEEEE"",""name"":""EEEEE"",""initial"":0},{""id"":""RRRRR"",""name"":""RRRRR"",""initial"":0}]}}";
                File.WriteAllText(Path.Combine(outDir, "project.janproj"), dummy, new UTF8Encoding(false));
            }

            // 2. KOPIUJ _statsSystem.json i avatarSystem dla kompatybilnosci (jesli istnieja)
            foreach (var name in new[] { "_statsSystem.json", "_avatarSystem.json", "avatarSystem.json", "_manifest.json" })
            {
                var src = Path.Combine(assetsDir, name);
                if (File.Exists(src))
                {
                    File.Copy(src, Path.Combine(outDir, name), true);
                    Console.WriteLine($"[BUNDLER] {name} copied");
                }
            }

            // 3. KOPIUJ day*.json 1:1 DO wwwroot/data - TO FIXUJE 404 na day1.json
            var files = Directory.GetFiles(assetsDir, "day*.json").OrderBy(f => f).ToArray();
            if (files.Length == 0)
            {
                Console.WriteLine($"[BUNDLER ERROR] Brak plikow day*.json w {assetsDir}");
                return 1;
            }

            var all = new JsonArray();
            foreach (var file in files)
            {
                var filename = Path.GetFileName(file);
                // kopiuj 1:1
                File.Copy(file, Path.Combine(outDir, filename), true);
                Console.WriteLine($"[BUNDLER] {filename} copied 1:1 -> wwwroot/data");

                var json = File.ReadAllText(file, Encoding.UTF8);
                var node = JsonNode.Parse(json);
                if (node == null) continue;

                if (node is JsonObject obj && obj["scenes"] is JsonArray scenes)
                {
                    foreach (var s in scenes) if (s!= null) all.Add(s.DeepClone());
                }
                else if (node is JsonArray arr)
                {
                    foreach (var s in arr) if (s!= null) all.Add(s.DeepClone());
                }
            }

            // 4. ZAPISZ scenes.json (aggregate) dla fallbacku
            var result = new JsonObject { ["scenes"] = all };
            var resultJson = result.ToJsonString(new JsonSerializerOptions
            {
                WriteIndented = true,
                Encoder = System.Text.Encodings.Web.JavaScriptEncoder.UnsafeRelaxedJsonEscaping
            });

            var outPath = Path.Combine(outDir, "scenes.json");
            File.WriteAllText(outPath, resultJson, new UTF8Encoding(false));
            Console.WriteLine($"[BUNDLER OK] Zapisano {all.Count} scen z {files.Length} plikow -> {outPath}");
            return 0;
        }
        catch (Exception ex)
        {
            Console.WriteLine($"[BUNDLER CRASH] {ex.GetType().Name}: {ex.Message}");
            Console.WriteLine(ex.StackTrace);
            return 1;
        }
    }
}