using System;
using System.IO;
using System.IO.Compression;
using System.Linq;
using System.Net;
using System.Net.Sockets;
using System.Threading;
using System.Windows.Forms;
using Microsoft.Web.WebView2.WinForms;
using Microsoft.Web.WebView2.Core;

class Program {
    static string GetMime(string path) {
        var ext = Path.GetExtension(path).ToLowerInvariant();
        return ext switch {
            ".html" => "text/html",
            ".js" => "application/javascript",
            ".css" => "text/css",
            ".json" => "application/json",
            ".png" => "image/png",
            ".jpg" or ".jpeg" => "image/jpeg",
            ".webp" => "image/webp",
            ".gif" => "image/gif",
            ".svg" => "image/svg+xml",
            ".wasm" => "application/wasm",
            ".woff" => "font/woff",
            ".woff2" => "font/woff2",
            ".ttf" => "font/ttf",
            ".ico" => "image/x-icon",
            _ => "application/octet-stream",
        };
    }

    static byte[] ReadAppendedZip(string exePath) {
        // Krok 5 single-file: edytor dokleja wwwroot.zip na koniec exe + footer [8 bajtow dlugosc][8 bajtow magia KIMONO01]
        // Launcher czyta go z samego siebie, nie szuka pliku obok.
        try {
            if (string.IsNullOrEmpty(exePath) || !File.Exists(exePath)) return null;
            using var fs = new FileStream(exePath, FileMode.Open, FileAccess.Read, FileShare.Read);
            if (fs.Length < 16) return null;
            fs.Seek(-16, SeekOrigin.End);
            var footer = new byte[16];
            if (fs.Read(footer, 0, 16) != 16) return null;
            var magic = System.Text.Encoding.ASCII.GetString(footer, 8, 8);
            if (magic != "KIMONO01") return null;
            long zipLen = BitConverter.ToInt64(footer, 0);
            if (zipLen <= 0 || zipLen > fs.Length - 16) return null;
            fs.Seek(-16 - zipLen, SeekOrigin.End);
            var zipBytes = new byte[zipLen];
            int read = 0;
            while (read < zipLen) {
                int n = fs.Read(zipBytes, read, (int)(zipLen - read));
                if (n <= 0) break;
                read += n;
            }
            if (read != zipLen) return null;
            // Sprawdz czy to na pewno zip z index.html
            using var ms = new MemoryStream(zipBytes);
            using var z = new System.IO.Compression.ZipArchive(ms, System.IO.Compression.ZipArchiveMode.Read);
            foreach (var e in z.Entries) {
                if (e.FullName.Replace('\\', '/').TrimStart('/').Equals("index.html", StringComparison.OrdinalIgnoreCase)) return zipBytes;
            }
            return null;
        } catch { return null; }
    }

    static string FindWwwrootZip() {
        var candidates = new[] {
            Path.Combine(AppContext.BaseDirectory, "wwwroot.zip"),
            Path.Combine(Path.GetDirectoryName(Application.ExecutablePath), "wwwroot.zip"),
            Path.Combine(Directory.GetCurrentDirectory(), "wwwroot.zip"),
            Path.Combine(Path.GetTempPath(), "KimonoGM", "wwwroot.zip"),
        };
        foreach (var p in candidates) if (File.Exists(p)) return p;
        try {
            foreach (var f in Directory.GetFiles(AppContext.BaseDirectory, "wwwroot.zip", SearchOption.AllDirectories)) {
                if (File.Exists(f)) return f;
            }
        } catch {}
        return null;
    }

    [STAThread]
    static void Main() {
        Application.EnableVisualStyles();
        Application.SetCompatibleTextRenderingDefault(false);

        var exeName = Path.GetFileNameWithoutExtension(Application.ExecutablePath);
        var tempRoot = Path.Combine(Path.GetTempPath(), "KimonoGM", exeName);
        var wwwroot = Path.Combine(tempRoot, "wwwroot");

        // Krok 5: najpierw sprobuj zipa doklejonego na koniec tego exe (single-file)
        // ProcessPath zawsze pokazuje prawdziwe Pobrane/q.exe, nawet w single-file
        var exePathForAppended = Environment.ProcessPath;
        if (string.IsNullOrEmpty(exePathForAppended)) exePathForAppended = Application.ExecutablePath;
        byte[] appendedZip = null;
        try { appendedZip = ReadAppendedZip(exePathForAppended); } catch {}
        // Jak jest appended zip, wypakuj go od razu do wwwroot i nie szukaj dalej
        if (appendedZip != null) {
            try {
                var marker2 = Path.Combine(tempRoot, "_ver.txt");
                var zipTime2 = appendedZip.Length.ToString() + "_" + new FileInfo(exePathForAppended).Length.ToString();
                bool need2 = true;
                if (File.Exists(marker2) && File.ReadAllText(marker2) == zipTime2 && Directory.Exists(wwwroot) && File.Exists(Path.Combine(wwwroot, "index.html"))) need2 = false;
                if (need2) {
                    if (Directory.Exists(wwwroot)) try { Directory.Delete(wwwroot, true); } catch {}
                    Directory.CreateDirectory(wwwroot);
                    using var ms2 = new MemoryStream(appendedZip);
                    using var z2 = new System.IO.Compression.ZipArchive(ms2, System.IO.Compression.ZipArchiveMode.Read);
                    foreach (var e2 in z2.Entries) {
                        var dest = Path.GetFullPath(Path.Combine(wwwroot, e2.FullName.Replace('/', Path.DirectorySeparatorChar)));
                        if (!dest.StartsWith(Path.GetFullPath(wwwroot), StringComparison.OrdinalIgnoreCase)) continue;
                        if (string.IsNullOrEmpty(e2.Name)) { Directory.CreateDirectory(dest); continue; }
                        Directory.CreateDirectory(Path.GetDirectoryName(dest));
                        e2.ExtractToFile(dest, true);
                    }
                    Directory.CreateDirectory(tempRoot);
                    File.WriteAllText(marker2, zipTime2);
                }
            } catch (Exception ex) {
                MessageBox.Show("Blad wypakowania doklejonego wwwroot.zip: " + ex.Message);
            }
        }

        var wwwrootZip = appendedZip != null ? null : FindWwwrootZip();

        try {
            if (!string.IsNullOrEmpty(wwwrootZip) && File.Exists(wwwrootZip)) {
                var marker = Path.Combine(tempRoot, "_ver.txt");
                var zipTime = File.GetLastWriteTimeUtc(wwwrootZip).Ticks.ToString() + "_" + new FileInfo(wwwrootZip).Length.ToString();
                var needExtract = true;
                if (File.Exists(marker) && File.ReadAllText(marker) == zipTime && Directory.Exists(wwwroot) && File.Exists(Path.Combine(wwwroot, "index.html"))) {
                    needExtract = false;
                }
                if (needExtract) {
                    if (Directory.Exists(wwwroot)) try { Directory.Delete(wwwroot, true); } catch {}
                    Directory.CreateDirectory(wwwroot);
                    ZipFile.ExtractToDirectory(wwwrootZip, wwwroot, true);
                    Directory.CreateDirectory(tempRoot);
                    File.WriteAllText(marker, zipTime);
                }
            }
        } catch (Exception ex) {
            MessageBox.Show("Blad wypakowania wwwroot.zip: " + ex.Message);
        }

        if (!File.Exists(Path.Combine(wwwroot, "index.html"))) {
            var fallback = new[] {
                Path.Combine(AppContext.BaseDirectory, "wwwroot"),
                Path.Combine(Path.GetDirectoryName(Application.ExecutablePath), "wwwroot"),
                Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), "Downloads", "wwwroot"),
            };
            foreach (var c in fallback) {
                if (File.Exists(Path.Combine(c, "index.html"))) { wwwroot = c; break; }
            }
        }

        if (!File.Exists(Path.Combine(wwwroot, "index.html"))) {
            MessageBox.Show("Brak wwwroot/index.html");
            return;
        }

        int port = 0;
        var tcp = new TcpListener(IPAddress.Loopback, 0);
        tcp.Start();
        port = ((IPEndPoint)tcp.LocalEndpoint).Port;
        tcp.Stop();

        var http = new HttpListener();
        http.Prefixes.Add($"http://localhost:{port}/");
        http.Start();

        ThreadPool.QueueUserWorkItem(_ => {
            while (http.IsListening) {
                try {
                    var ctx = http.GetContext();
                    var resp = ctx.Response;
                    var urlPath = ctx.Request.Url.AbsolutePath.TrimStart('/');
                    if (string.IsNullOrEmpty(urlPath)) urlPath = "index.html";
                    var filePath = Path.GetFullPath(Path.Combine(wwwroot, urlPath.Replace('/', Path.DirectorySeparatorChar)));
                    if (!filePath.StartsWith(Path.GetFullPath(wwwroot), StringComparison.OrdinalIgnoreCase)) {
                        filePath = Path.Combine(wwwroot, "index.html");
                    }
                    if (!File.Exists(filePath) && !urlPath.Contains(".")) {
                        filePath = Path.Combine(wwwroot, "index.html");
                    }
                    if (File.Exists(filePath)) {
                        var bytes = File.ReadAllBytes(filePath);
                        resp.ContentType = GetMime(filePath);
                        resp.ContentLength64 = bytes.Length;
                        resp.OutputStream.Write(bytes, 0, bytes.Length);
                    } else {
                        resp.StatusCode = 404;
                        var notFound = System.Text.Encoding.UTF8.GetBytes("404 Not Found");
                        resp.ContentLength64 = notFound.Length;
                        resp.OutputStream.Write(notFound, 0, notFound.Length);
                    }
                    resp.OutputStream.Close();
                } catch { }
            }
        });

        var baseDir = AppContext.BaseDirectory;
        var iconPath = Path.Combine(baseDir, "icon.ico");
        if (!File.Exists(iconPath)) iconPath = Path.Combine(Path.GetDirectoryName(Application.ExecutablePath), "icon.ico");
        System.Drawing.Icon icon = null;
        if (File.Exists(iconPath)) { try { icon = new System.Drawing.Icon(iconPath); } catch {} }

        var form = new Form() { Width=1280, Height=800, Text="KimonoGM Game", StartPosition=FormStartPosition.CenterScreen, Icon = icon };
        var webView = new WebView2() { Dock=DockStyle.Fill };
        form.Controls.Add(webView);
        form.FormClosed += (s,e) => { try { http.Stop(); } catch {} Environment.Exit(0); };
        form.Load += async (s,e) => {
            try {
                var webView2Temp = Path.Combine(Path.GetTempPath(), "KimonoGM", "WebView2", exeName);
                Directory.CreateDirectory(webView2Temp);
                var env = await CoreWebView2Environment.CreateAsync(null, webView2Temp);
                await webView.EnsureCoreWebView2Async(env);
            } catch {
                await webView.EnsureCoreWebView2Async();
            }
            webView.CoreWebView2.Navigate($"http://localhost:{port}/");
        };
        Application.Run(form);
    }
}