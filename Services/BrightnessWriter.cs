using System.ComponentModel;
using BrightSlide.Models;

namespace BrightSlide.Services;

/// <summary>
/// Pushes brightness to hardware as fast as the hardware allows: one
/// WMI/DDC-CI call in flight per display at all times, and the instant it
/// finishes, the newest slider value (if any) is sent immediately — no
/// polling interval, no debounce wait. Throughput is bounded purely by how
/// long the monitor/driver takes to answer each call (DDC/CI monitors are
/// commonly ~50-70ms per call).
///
/// DDC/CI monitors silently drop commands that arrive in quick succession, so
/// once the slider goes idle the writer reads the value back and re-sends the
/// final value until the panel actually reports it.
/// </summary>
public sealed class BrightnessWriter
{
    private const int VerifyAttempts = 3;
    private const int SettleSteps = 6;
    private const int SettleStepMs = 20;

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
        lock (_gate)
        {
            if (_isWriting) return;
            if (_device.Brightness == _lastSent) return;
            _isWriting = true;
        }

        Task.Run(PumpLoop);
    }

    private void PumpLoop()
    {
        try
        {
            int attemptsLeft = VerifyAttempts;
            while (true)
            {
                int target = _device.Brightness;
                if (target != _lastSent)
                {
                    _lastSent = target;
                    _device.Controller.SetBrightness(target);
                    attemptsLeft = VerifyAttempts;
                    continue;
                }

                // The external controller remaps 1-8% to 9%, so the read-back can't be compared there.
                if (attemptsLeft == 0 || target is >= 1 and <= 8)
                    break;

                if (!WaitForIdle(target))
                    continue;

                if (ReadBackMatches(target))
                    break;

                attemptsLeft--;
                _device.Controller.SetBrightness(target);
            }
        }
        finally
        {
            lock (_gate) { _isWriting = false; }
            // A value that arrived after the last check but before the flag cleared was skipped by TryPump.
            TryPump();
        }
    }

    // Returns false as soon as a newer slider value shows up.
    private bool WaitForIdle(int target)
    {
        for (int i = 0; i < SettleSteps; i++)
        {
            Thread.Sleep(SettleStepMs);
            if (_device.Brightness != target) return false;
        }
        return true;
    }

    private bool ReadBackMatches(int target)
    {
        try { return Math.Abs(_device.Controller.GetBrightness() - target) <= 1; }
        catch { return true; }
    }
}
