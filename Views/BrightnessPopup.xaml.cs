using System.Collections.ObjectModel;
using System.ComponentModel;
using System.Runtime.CompilerServices;
using System.Windows;
using System.Windows.Threading;
using BrightSlide.Models;

namespace BrightSlide.Views;

public partial class BrightnessPopup : Window
{
    private readonly PopupViewModel _viewModel;
    private readonly Dictionary<string, DispatcherTimer> _debounceTimers = new();

    public BrightnessPopup(IReadOnlyList<DisplayDevice> devices)
    {
        InitializeComponent();
        _viewModel = new PopupViewModel(devices);
        DataContext = _viewModel;

        foreach (var device in devices)
            device.PropertyChanged += OnDeviceBrightnessChanged;
    }

    private void OnDeviceBrightnessChanged(object? sender, PropertyChangedEventArgs e)
    {
        if (e.PropertyName != nameof(DisplayDevice.Brightness) || sender is not DisplayDevice device)
            return;

        if (!_debounceTimers.TryGetValue(device.Id, out var timer))
        {
            timer = new DispatcherTimer { Interval = TimeSpan.FromMilliseconds(100) };
            timer.Tick += (_, _) =>
            {
                timer!.Stop();
                var percent = device.Brightness;
                Task.Run(() => device.Controller.SetBrightness(percent));
            };
            _debounceTimers[device.Id] = timer;
        }

        timer.Stop();
        timer.Start();
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

internal sealed class PopupViewModel : INotifyPropertyChanged
{
    public ObservableCollection<DisplayDevice> Devices { get; }

    public Visibility HasNoDevicesVisibility => Devices.Count == 0 ? Visibility.Visible : Visibility.Collapsed;

    public PopupViewModel(IReadOnlyList<DisplayDevice> devices)
    {
        Devices = new ObservableCollection<DisplayDevice>(devices);
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    private void OnPropertyChanged([CallerMemberName] string? name = null) =>
        PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(name));
}
