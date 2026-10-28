using Microsoft.Win32;

namespace BrightSlide.Services;

/// <summary>
/// Registers/unregisters BrightSlide in the per-user "Run" key so it can
/// launch automatically at sign-in. Purely a user preference toggle.
/// </summary>
public static class StartupManager
{
    private const string RunKeyPath = @"Software\Microsoft\Windows\CurrentVersion\Run";
    private const string ValueName = "BrightSlide";

    public static bool IsEnabled
    {
        get
        {
            using var key = Registry.CurrentUser.OpenSubKey(RunKeyPath, writable: false);
            return key?.GetValue(ValueName) is string existing &&
                   existing.Equals(GetExecutablePath(), StringComparison.OrdinalIgnoreCase);
        }
        set
        {
            using var key = Registry.CurrentUser.OpenSubKey(RunKeyPath, writable: true)
                ?? Registry.CurrentUser.CreateSubKey(RunKeyPath, writable: true)!;

            if (value)
                key.SetValue(ValueName, GetExecutablePath());
            else
                key.DeleteValue(ValueName, throwOnMissingValue: false);
        }
    }

    private static string GetExecutablePath() =>
        Environment.ProcessPath ?? throw new InvalidOperationException("실행 파일 경로를 확인할 수 없습니다.");
}
