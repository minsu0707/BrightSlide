using System.Runtime.InteropServices;
using BrightSlide.Interop;
using static BrightSlide.Interop.MonitorNativeMethods;

namespace BrightSlide.Services;

/// <summary>
/// Controls the brightness of an external monitor over DDC/CI
/// (VCP code 0x10 = luminance) via the Monitor Configuration API.
/// Requires the monitor and its cable (HDMI/DP) to support DDC/CI.
/// </summary>
public sealed class ExternalDisplayController : IDisplayController
{
    private const byte VcpBrightness = 0x10;

    private readonly IntPtr _handle;
    private readonly uint _maxValue;

    public string Id { get; }
    public string Name { get; }

    private ExternalDisplayController(IntPtr handle, uint maxValue, string id, string name)
    {
        _handle = handle;
        _maxValue = maxValue == 0 ? 100 : maxValue;
        Id = id;
        Name = name;
    }

    public int GetBrightness()
    {
        if (!GetVCPFeatureAndVCPFeatureReply(_handle, VcpBrightness, IntPtr.Zero, out uint current, out _))
            return 0;
        return (int)Math.Round(current * 100.0 / _maxValue);
    }

    public void SetBrightness(int percent)
    {
        percent = Math.Clamp(percent, 0, 100);

        // Some DDC/CI monitors have a non-monotonic dimming curve near the low
        // end (1-8% can render visibly darker than 0%). Skip that dead zone
        // rather than send a value the panel renders "wrong".
        if (percent is >= 1 and <= 8)
            percent = 9;

        uint raw = (uint)Math.Round(percent * _maxValue / 100.0);
        SetVCPFeature(_handle, VcpBrightness, raw);
    }

    public static List<ExternalDisplayController> DiscoverAll()
    {
        var controllers = new List<ExternalDisplayController>();
        var friendlyNamesByHardwareId = MonitorFriendlyNameResolver.LoadByHardwareId();
        int index = 0;

        bool EnumProc(IntPtr hMonitor, IntPtr hdcMonitor, ref RECT lprcMonitor, IntPtr dwData)
        {
            if (!GetNumberOfPhysicalMonitorsFromHMONITOR(hMonitor, out uint count) || count == 0)
                return true;

            var physicalMonitors = new PHYSICAL_MONITOR[count];
            if (!GetPhysicalMonitorsFromHMONITOR(hMonitor, count, physicalMonitors))
                return true;

            var modelName = TryResolveModelName(hMonitor, friendlyNamesByHardwareId);

            foreach (var pm in physicalMonitors)
            {
                // Probe DDC/CI support; monitors (or cables) without it fail here.
                if (GetVCPFeatureAndVCPFeatureReply(pm.hPhysicalMonitor, VcpBrightness, IntPtr.Zero, out _, out uint max))
                {
                    index++;
                    var label = BuildLabel(index, modelName ?? NullIfBlank(pm.szPhysicalMonitorDescription));
                    controllers.Add(new ExternalDisplayController(pm.hPhysicalMonitor, max, $"ddcci-{index}", label));
                }
                else
                {
                    DestroyPhysicalMonitor(pm.hPhysicalMonitor);
                }
            }
            return true;
        }

        EnumDisplayMonitors(IntPtr.Zero, IntPtr.Zero, EnumProc, IntPtr.Zero);
        return controllers;
    }

    private static string BuildLabel(int index, string? description) =>
        description is null ? $"외부 모니터 {index}" : $"외부 모니터 {index} ({description})";

    private static string? NullIfBlank(string value) => string.IsNullOrWhiteSpace(value) ? null : value;

    /// <summary>
    /// Resolves hMonitor -&gt; its Win32 display device name (\\.\DISPLAYn) -&gt;
    /// its hardware-id segment ("SKG2712") -&gt; the real model name from EDID,
    /// instead of showing Windows' generic DDC/CI description string.
    /// </summary>
    private static string? TryResolveModelName(IntPtr hMonitor, Dictionary<string, string> friendlyNamesByHardwareId)
    {
        var info = new MONITORINFOEX { cbSize = Marshal.SizeOf<MONITORINFOEX>() };
        if (!GetMonitorInfo(hMonitor, ref info))
            return null;

        var device = new DISPLAY_DEVICE { cb = Marshal.SizeOf<DISPLAY_DEVICE>() };
        if (!EnumDisplayDevices(info.szDevice, 0, ref device, 0))
            return null;

        // DeviceID looks like "MONITOR\SKG2712\{4d36e96e-...}\0000"; the second
        // segment is the hardware id shared with WmiMonitorID's InstanceName.
        var segments = device.DeviceID.Split('\\');
        if (segments.Length < 2)
            return null;

        return friendlyNamesByHardwareId.TryGetValue(segments[1], out var name) ? name : null;
    }

    public static void ReleaseAll(IEnumerable<ExternalDisplayController> controllers)
    {
        foreach (var c in controllers)
            DestroyPhysicalMonitor(c._handle);
    }

    private static void DestroyPhysicalMonitor(IntPtr handle)
    {
        var arr = new[] { new PHYSICAL_MONITOR { hPhysicalMonitor = handle, szPhysicalMonitorDescription = string.Empty } };
        DestroyPhysicalMonitors(1, arr);
    }
}
