using System.Windows;
using BrightSlide.Models;
using BrightSlide.Services;

namespace BrightSlide.Views;

public partial class BrightnessPopup : Window
{
    private readonly PopupViewModel _viewModel;
    private readonly List<BrightnessWriter> _writers = new();

    public BrightnessPopup(IReadOnlyList<DisplayDevice> devices)
    {
        InitializeComponent();
        _viewModel = new PopupViewModel(devices);
        DataContext = _viewModel;

        foreach (var device in devices)
            _writers.Add(new BrightnessWriter(device));
    }

    private void OnDeactivated(object? sender, EventArgs e) => Hide();

    public void ShowNear(Point screenPoint)
    {
        Show();
        UpdateLayout();
        Left = screenPoint.X - ActualWidth;
        Top = screenPoint.Y - ActualHeight - 12;
        Activate();
    }
}
