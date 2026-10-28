using System.Windows;
using System.Windows.Controls;
using BrightSlide.Interop;
using BrightSlide.Models;
using BrightSlide.Services;
using BrightSlide.Views;
using H.NotifyIcon;

namespace BrightSlide;

public partial class App : Application
{
    private TaskbarIcon? _taskbarIcon;
    private BrightnessPopup? _popup;
    private DisplayDiscoveryService? _discoveryService;
    private List<DisplayDevice> _devices = new();

    protected override void OnStartup(StartupEventArgs e)
    {
        base.OnStartup(e);

        _discoveryService = new DisplayDiscoveryService();
        _devices = _discoveryService.DiscoverAll();

        var contextMenu = new ContextMenu();

        var startupItem = new MenuItem { Header = "시작 시 자동 실행", IsCheckable = true, IsChecked = StartupManager.IsEnabled };
        startupItem.Click += (_, _) => StartupManager.IsEnabled = startupItem.IsChecked;
        contextMenu.Items.Add(startupItem);

        contextMenu.Items.Add(new Separator());

        var exitItem = new MenuItem { Header = "종료" };
        exitItem.Click += (_, _) => Shutdown();
        contextMenu.Items.Add(exitItem);

        _taskbarIcon = new TaskbarIcon
        {
            Icon = LoadTrayIcon(),
            ToolTipText = "BrightSlide — 모니터 밝기 조절",
            ContextMenu = contextMenu
        };
        _taskbarIcon.TrayLeftMouseUp += OnTrayLeftMouseUp;
        _taskbarIcon.ForceCreate(enablesEfficiencyMode: false);

        _popup = new BrightnessPopup(_devices);
    }

    private void OnTrayLeftMouseUp(object? sender, RoutedEventArgs e)
    {
        if (_popup is null) return;

        if (_popup.IsVisible)
        {
            _popup.Hide();
            return;
        }

        CursorNativeMethods.GetCursorPos(out var cursor);
        _popup.ShowNear(new Point(cursor.X, cursor.Y));
    }

    private static System.Drawing.Icon LoadTrayIcon()
    {
        var assembly = System.Reflection.Assembly.GetExecutingAssembly();
        using var stream = assembly.GetManifestResourceStream("BrightSlide.Assets.brightslide.ico");
        return stream is not null
            ? new System.Drawing.Icon(stream)
            : System.Drawing.SystemIcons.Application;
    }

    protected override void OnExit(ExitEventArgs e)
    {
        _taskbarIcon?.Dispose();
        _discoveryService?.ReleaseExternalHandles();
        base.OnExit(e);
    }
}
