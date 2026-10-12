using LiteDB;
using System.Text.Json;

var manifestPath = args[0];
var outDb = args[1];
var json = File.ReadAllText(manifestPath);
using var doc = JsonDocument.Parse(json);

if (File.Exists(outDb)) {
    try { File.Delete(outDb); } catch { Thread.Sleep(500); try{File.Delete(outDb);}catch{} }
}
using var db = new LiteDatabase(outDb);

BsonDocument JsonToBson(string t) => LiteDB.JsonSerializer.Deserialize(t).AsDocument;

void Upsert(string name, JsonElement el)
{
    if (el.ValueKind==JsonValueKind.Undefined||el.ValueKind==JsonValueKind.Null) return;
    var col=db.GetCollection<BsonDocument>(name);
    col.EnsureIndex("_id");
    if(el.ValueKind==JsonValueKind.Array){
        foreach(var e in el.EnumerateArray()){
            var bson=JsonToBson(e.GetRawText());
            if(!bson.ContainsKey("_id")||bson["_id"].IsNull||bson["_id"].AsString=="")
                bson["_id"]=ObjectId.NewObjectId().ToString();
            col.Upsert(bson);
        }
    } else if(el.ValueKind==JsonValueKind.Object){
        var bson=JsonToBson(el.GetRawText());
        if(!bson.ContainsKey("_id")) bson["_id"]="project";
        col.Upsert(bson);
    }
    Console.WriteLine($"[packer] {name}: {col.Count()}");
}

Upsert("scenes", doc.RootElement.GetProperty("scenes"));
if(doc.RootElement.TryGetProperty("avatars", out var av)) Upsert("avatars", av);
if(doc.RootElement.TryGetProperty("sounds", out var so)) Upsert("sounds", so);
if(doc.RootElement.TryGetProperty("stats", out var st)) Upsert("stats", st);
if(doc.RootElement.TryGetProperty("variables", out var va)) Upsert("variables", va);
if(doc.RootElement.TryGetProperty("project", out var pr)&&pr.ValueKind!=JsonValueKind.Null) Upsert("project", pr);

// --- BLOBS v5: osobna kolekcja meta + filestorage z bezpiecznym ID ---
var colMeta = db.GetCollection<BsonDocument>("blobs");
colMeta.EnsureIndex("_id");
var fs = db.FileStorage;
int ok=0;
if(doc.RootElement.TryGetProperty("files", out var files)){
    foreach(var f in files.EnumerateArray()){
        var rel = f.GetProperty("relPath").GetString()!;
        var full = f.GetProperty("fullPath").GetString()!;
        var ext = f.GetProperty("ext").GetString()!;
        if(!File.Exists(full)) { Console.WriteLine($"missing {full}"); continue; }

        var safeId = rel.Replace("/", "_").Replace("\\", "_"); // images_bg.png
        var fileInfo = new FileInfo(full);

        // 1. meta
        colMeta.Upsert(new BsonDocument{
            ["_id"] = rel,
            ["fileId"] = safeId,
            ["ext"] = ext,
            ["size"] = fileInfo.Length
        });

        // 2. binary w filestorage
        using var stream = File.OpenRead(full);
        // usuń jeśli istnieje
        try { fs.Delete(safeId); } catch {}
        fs.Upload(safeId, safeId, stream);
        Console.WriteLine($" -> blob {rel} => {safeId} ({fileInfo.Length} bytes)");
        ok++;
    }
}
Console.WriteLine($"[packer] blobs meta: {colMeta.Count()}, filestorage: {fs.FindAll().Count()} / {ok} uploaded");
Console.WriteLine($"[packer] ZAPISANO -> {outDb}");