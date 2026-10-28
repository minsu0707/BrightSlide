using System.ComponentModel;
using BrightSlide.Models;

namespace BrightSlide.Services;

/// <summary>
/// Pushes brightness to hardware as fast as the hardware allows: one
/// WMI/DDC-CI call in flight per display at all times, and the instant it
/// finishes, the newest slider value (if any) is sent immediately — no
/// polling interval, no debounce wait. Throughput is bounded purely by how
/// long the monitor/driver takes to answer each call (DDC/CI monitors are
/// commonly ~50-70ms per call; this keeps the gap between calls at ~0).
/// </summary>
public sealed class BrightnessWriter
{
    private readonly DisplayDevice _device;
    private readonly object _gate = new();
    private int _lastSent;
    private bool _isWriting;

    public BrightnessWriter(DisplayDevice device)
    {
        _device = device;
        _lastSent = device.Brightness;
        device.PropertyChanged += OnBrightnessChanged;
    }

    private void OnBrightnessChanged(object? sender, PropertyChangedEventArgs e)
    {
        if (e.PropertyName == nameof(DisplayDevice.Brightness))
            TryPump();
    }

    private void TryPump()
    {
        int valueToSend;
        lock (_gate)
        {
            if (_isWriting) return;
            valueToSend = _device.Brightness;
            if (valueToSend == _lastSent) return;
            _isWriting = true;
            _lastSent = valueToSend;
        }

        Task.Run(() =>
        {
            try { _device.Controller.SetBrightness(valueToSend); }
            finally
            {
                lock (_gate) { _isWriting = false; }
                TryPump();
            }
        });
    }
}
