package main

import (
  "net/http"
  "net/http/httptest"
  "os"
  "path/filepath"
  "testing"

  "rewinged/models"
)

// A declared hash that no test payload actually matches. Nothing in this file
// depends on it being correct - what matters is the file the cache is named
// after either exists complete or does not exist at all.
const declaredSha = "5b04177b0ce3a6b56a744a6f6910d506f295c0463415e69c023df77a0c07f14e"

func testInstaller(url string, sha string) models.API_InstallerInterface {
  return &models.API_Installer_1_6_0{
    Architecture:    "x64",
    InstallerType:   "exe",
    InstallerUrl:    url,
    InstallerSha256: sha,
  }
}

// cacheEntries lists what is in the internalize path, so a test can assert on
// leftovers rather than only on the entry it expected.
func cacheEntries(t *testing.T, dir string) []os.DirEntry {
  t.Helper()
  entries, err := os.ReadDir(dir)
  if err != nil {
    t.Fatalf("reading %s: %v", dir, err)
  }
  return entries
}

// A download that does not finish must leave nothing behind. The destination is
// named by the declared hash, so a leftover - empty or truncated - is read as a
// completed download by the next start: the URL is rewritten to point at this
// instance and the source then serves those bytes to every client, on a 200,
// with no error logged after the run that failed.
//
// Leaving nothing also means the next start retries, which a leftover prevents.
func TestInternalizeInstallersLeavesNothingOnFailure(t *testing.T) {
  // A body shorter than the Content-Length it promises, with the connection
  // closed afterwards, so io.Copy fails partway and the destination would hold
  // a truncated file rather than an empty one.
  truncating := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Length", "4096")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("partial payload, then nothing"))
    if f, ok := w.(http.Flusher); ok {
      f.Flush()
    }
    // Returning without writing the promised remainder ends the response early.
  }))
  defer truncating.Close()

  serverError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "no", http.StatusInternalServerError)
  }))
  defer serverError.Close()

  // A server that is listening at the point the URL is built and closed before
  // the request, which fails in the transport rather than with a status.
  unreachable := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
  unreachableURL := unreachable.URL + "/installer.exe"
  unreachable.Close()

  tests := []struct {
    name string
    url  string
  }{
    {name: "transport error", url: unreachableURL},
    {name: "non-2xx status", url: serverError.URL + "/installer.exe"},
    {name: "body truncated mid-copy", url: truncating.URL + "/installer.exe"},
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      dir := t.TempDir()

      internalizeInstallers("Contoso.Probe", "1.0.0",
        []models.API_InstallerInterface{testInstaller(tt.url, declaredSha)},
        dir, []string{})

      entries := cacheEntries(t, dir)
      for _, e := range entries {
        info, err := e.Info()
        size := int64(-1)
        if err == nil {
          size = info.Size()
        }
        t.Errorf("failed download left %q behind (%d bytes); the next start would serve it as a cache hit", e.Name(), size)
      }

      // A failure must not mark the installer internalized either. Doing so
      // rewrites the served URL to an instance that has nothing to serve.
      if models.InternalizedInstallers.Get(declaredSha) {
        t.Errorf("installer was marked internalized despite the download failing")
      }
    })
  }
}

// The success path has to keep working: the payload lands under the declared
// hash, complete, and the installer is marked internalized so its URL is
// rewritten.
func TestInternalizeInstallersStoresACompleteDownload(t *testing.T) {
  payload := []byte("a complete installer payload")

  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    _, _ = w.Write(payload)
  }))
  defer server.Close()

  dir := t.TempDir()
  sha := "aaaa1111bbbb2222cccc3333dddd4444eeee5555ffff6666aaaa7777bbbb8888"

  internalizeInstallers("Contoso.Probe", "1.0.0",
    []models.API_InstallerInterface{testInstaller(server.URL+"/installer.exe", sha)},
    dir, []string{})

  got, err := os.ReadFile(filepath.Join(dir, sha))
  if err != nil {
    t.Fatalf("expected the installer cached under its declared hash: %v", err)
  }
  if string(got) != string(payload) {
    t.Errorf("cached %q, want %q", got, payload)
  }
  if !models.InternalizedInstallers.Get(sha) {
    t.Error("a completed download should mark the installer internalized")
  }
  if n := len(cacheEntries(t, dir)); n != 1 {
    t.Errorf("expected exactly one cache entry, got %d", n)
  }

  // Staging through a temporary file must not change the mode the entry ends up
  // with - CreateTemp defaults to 0600, and anything outside this process that
  // reads the directory would start failing.
  info, err := os.Stat(filepath.Join(dir, sha))
  if err != nil {
    t.Fatalf("stat on the cached entry: %v", err)
  }
  if got := info.Mode().Perm(); got != 0644 {
    t.Errorf("cached entry has mode %o, want %o", got, 0644)
  }
}

// An entry that is already present is a cache hit: it must not be re-downloaded
// and must not be overwritten, even when the upstream is reachable.
func TestInternalizeInstallersKeepsAnExistingEntry(t *testing.T) {
  var requests int
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    requests++
    _, _ = w.Write([]byte("freshly downloaded"))
  }))
  defer server.Close()

  dir := t.TempDir()
  sha := "1111aaaa2222bbbb3333cccc4444dddd5555eeee6666ffff7777aaaa8888bbbb"
  existing := []byte("already cached")

  if err := os.WriteFile(filepath.Join(dir, sha), existing, 0o666); err != nil {
    t.Fatalf("seeding the cache: %v", err)
  }

  internalizeInstallers("Contoso.Probe", "1.0.0",
    []models.API_InstallerInterface{testInstaller(server.URL+"/installer.exe", sha)},
    dir, []string{})

  if requests != 0 {
    t.Errorf("made %d request(s); an existing entry should not be re-downloaded", requests)
  }
  got, err := os.ReadFile(filepath.Join(dir, sha))
  if err != nil {
    t.Fatalf("reading the cached entry: %v", err)
  }
  if string(got) != string(existing) {
    t.Errorf("cache entry was overwritten: got %q, want %q", got, existing)
  }
  if !models.InternalizedInstallers.Get(sha) {
    t.Error("an existing entry should still mark the installer internalized")
  }
}
