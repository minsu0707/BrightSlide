using System.Windows;
using BrightSlide.Models;
using BrightSlide.Services;

namespace BrightSlide.Views;

public partial class BrightnessPopup : Window
{
    // Clicking a tray icon that lives in the Windows "hidden icons" overflow
    // closes that overflow flyout first, which briefly shuffles window focus.
    // That shuffle lands right as this popup opens and fires a spurious
    // Deactivated, closing it the instant it appears. Ignore deactivation for
    // a brief grace period after showing so only a real "clicked elsewhere"
    // closes the popup.
    private static readonly TimeSpan DeactivateGracePeriod = TimeSpan.FromMilliseconds(250);

    private readonly PopupViewModel _viewModel;
    private readonly List<BrightnessWriter> _writers = new();
    private DateTime _shownAtUtc;

    public BrightnessPopup(IReadOnlyList<DisplayDevice> devices)
    {
        InitializeComponent();
        _viewModel = new PopupViewModel(devices);
        DataContext = _viewModel;

        foreach (var device in devices)
            _writers.Add(new BrightnessWriter(device));
    }

    private void OnDeactivated(object? sender, EventArgs e)
    {
        if (DateTime.UtcNow - _shownAtUtc < DeactivateGracePeriod)
            return;
        Hide();
    }

    public void ShowNear(Point screenPoint)
    {
        Show();
        UpdateLayout();
        Left = screenPoint.X - ActualWidth;
        Top = screenPoint.Y - ActualHeight - 12;
        Activate();
        _shownAtUtc = DateTime.UtcNow;
    }
}
