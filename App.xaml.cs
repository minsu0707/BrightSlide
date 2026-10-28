using System.Runtime.InteropServices;
using System.Windows;
using System.Windows.Controls;
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
        var exitItem = new MenuItem { Header = "종료" };
        exitItem.Click += (_, _) => Shutdown();
        contextMenu.Items.Add(exitItem);

        _taskbarIcon = new TaskbarIcon
        {
            Icon = System.Drawing.SystemIcons.Application,
            ToolTipText = "BrightSlide — 모니터 밝기 조절",
            ContextMenu = contextMenu
        };
        _taskbarIcon.TrayLeftMouseUp += OnTrayLeftMouseUp;

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

        GetCursorPos(out var cursor);
        _popup.ShowNear(new Point(cursor.X, cursor.Y));
    }

    protected override void OnExit(ExitEventArgs e)
    {
        _taskbarIcon?.Dispose();
        _discoveryService?.ReleaseExternalHandles();
        base.OnExit(e);
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct POINT
    {
        public int X;
        public int Y;
    }

    [DllImport("user32.dll")]
    private static extern bool GetCursorPos(out POINT lpPoint);
}
