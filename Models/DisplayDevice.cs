using System.ComponentModel;
using System.Runtime.CompilerServices;
using BrightSlide.Services;

namespace BrightSlide.Models;

public enum DisplayKind
{
    Internal,
    External
}

public sealed class DisplayDevice : INotifyPropertyChanged
{
    private int _brightness;

    public required string Id { get; init; }
    public required string Name { get; init; }
    public required DisplayKind Kind { get; init; }
    public required IDisplayController Controller { get; init; }

    public int Brightness
    {
        get => _brightness;
        set
        {
            if (_brightness == value) return;
            _brightness = value;
            OnPropertyChanged();
        }
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    private void OnPropertyChanged([CallerMemberName] string? name = null) =>
        PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(name));
}
