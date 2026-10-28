using System.Collections.ObjectModel;
using System.ComponentModel;
using System.Runtime.CompilerServices;
using System.Windows;
using BrightSlide.Models;

namespace BrightSlide.Views;

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
