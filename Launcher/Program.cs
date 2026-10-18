using System;
using System.IO;
using System.Linq;
using System.Net;
using System.Net.Sockets;
using System.Threading;
using System.Windows.Forms;
using Microsoft.Web.WebView2.WinForms;

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

    [STAThread]
    static void Main() {
        Application.EnableVisualStyles();
        Application.SetCompatibleTextRenderingDefault(false);
        var baseDir = AppContext.BaseDirectory;
        var wwwroot = Path.Combine(baseDir, "wwwroot");
        if (!Directory.Exists(wwwroot) || !File.Exists(Path.Combine(wwwroot, "index.html"))) {
            var candidates = new[] {
                Path.Combine(baseDir, "wwwroot"),
                Path.Combine(baseDir, "..", "dist", "wwwroot"),
                Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "wwwroot"),
                Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), "Downloads", "wwwroot"),
            };
            foreach (var c in candidates) if (Directory.Exists(c) && File.Exists(Path.Combine(c, "index.html"))) { wwwroot = c; break; }
            if (!File.Exists(Path.Combine(wwwroot, "index.html"))) {
                try {
                    var dl = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), "Downloads");
                    if (Directory.Exists(dl)) {
                        foreach (var dir in Directory.GetDirectories(dl, "*-EXE-*").OrderByDescending(d => d)) {
                            var p = Path.Combine(dir, "wwwroot");
                            if (File.Exists(Path.Combine(p, "index.html"))) { wwwroot = p; break; }
                        }
                    }
                } catch {}
            }
            if (!File.Exists(Path.Combine(wwwroot, "index.html"))) {
                try {
                    foreach (var dir in Directory.GetDirectories(baseDir)) {
                        var p = Path.Combine(dir, "wwwroot");
                        if (File.Exists(Path.Combine(p, "index.html"))) { wwwroot = p; break; }
                        var p2 = Path.Combine(dir, "wwwroot", "wwwroot");
                        if (File.Exists(Path.Combine(p2, "index.html"))) { wwwroot = p2; break; }
                    }
                } catch {}
            }
        }

        if (!File.Exists(Path.Combine(wwwroot, "index.html"))) {
            MessageBox.Show("Brak wwwroot/index.html w " + wwwroot + "\nSzukano w:\n" + baseDir + "/wwwroot\nDownloads/wwwroot\nDownloads/*-EXE-*/wwwroot");
            return;
        }

        // znajdz wolny port
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
                    var req = ctx.Request;
                    var resp = ctx.Response;
                    var urlPath = req.Url.AbsolutePath.TrimStart('/');
                    if (string.IsNullOrEmpty(urlPath)) urlPath = "index.html";
                    // zabezpiecz przed .. 
                    var filePath = Path.GetFullPath(Path.Combine(wwwroot, urlPath.Replace('/', Path.DirectorySeparatorChar)));
                    if (!filePath.StartsWith(Path.GetFullPath(wwwroot), StringComparison.OrdinalIgnoreCase)) {
                        filePath = Path.Combine(wwwroot, "index.html");
                    }
                    if (!File.Exists(filePath)) {
                        // SPA fallback - jesli plik nie istnieje a nie ma rozszerzenia, zwroc index.html
                        if (!Path.HasExtension(filePath) || !File.Exists(filePath)) {
                            // sprobuj czy to nie jest _framework, css itp - jesli nie ma, to index.html
                            if (!File.Exists(filePath) && !urlPath.Contains(".")) {
                                filePath = Path.Combine(wwwroot, "index.html");
                            }
                        }
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

        var form = new Form() { Width=1280, Height=800, Text="Janusz Simulator 97 - Manieczki", StartPosition=FormStartPosition.CenterScreen };
        var webView = new WebView2() { Dock=DockStyle.Fill };
        form.Controls.Add(webView);
        form.FormClosed += (s,e) => { try { http.Stop(); } catch {} Environment.Exit(0); };
        form.Load += async (s,e) => {
            await webView.EnsureCoreWebView2Async();
            webView.CoreWebView2.Navigate($"http://localhost:{port}/");
        };
        Application.Run(form);
    }
}