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

    static string FindWwwrootZip() {
        var candidates = new[] {
            Path.Combine(AppContext.BaseDirectory, "wwwroot.zip"),
            Path.Combine(Path.GetDirectoryName(Application.ExecutablePath), "wwwroot.zip"),
            Path.Combine(Directory.GetCurrentDirectory(), "wwwroot.zip"),
            Path.Combine(Path.GetTempPath(), "KimonoGM", "wwwroot.zip"),
        };
        // w single file mode AppContext.BaseDirectory to temp .net folder
        foreach (var p in candidates) if (File.Exists(p)) return p;
        // szukaj rekurencyjnie w baseDir
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

        // znajdz zip
        var wwwrootZip = FindWwwrootZip();

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
            MessageBox.Show("Blad wypakowania wwwroot.zip: " + ex.Message + "\nZip: " + wwwrootZip + "\nTemp: " + tempRoot);
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
            MessageBox.Show("Brak wwwroot/index.html\nSzukano w:\n" + wwwroot + "\nZip: " + (wwwrootZip ?? "nie znaleziono") + "\nBaseDir: " + AppContext.BaseDirectory + "\nExeDir: " + Path.GetDirectoryName(Application.ExecutablePath));
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
                        var notFound = System.Text.Encoding.UTF8.GetBytes("404 Not Found: " + urlPath);
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