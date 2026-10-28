using System.Runtime.InteropServices;

namespace BrightSlide.Services;

/// <summary>
/// Controls the brightness of an external monitor over DDC/CI
/// (VCP code 0x10 = luminance) via dxva2.dll / Monitor Configuration API.
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
        uint raw = (uint)Math.Round(percent * _maxValue / 100.0);
        SetVCPFeature(_handle, VcpBrightness, raw);
    }

    public static List<ExternalDisplayController> DiscoverAll()
    {
        var controllers = new List<ExternalDisplayController>();
        int index = 0;

        bool EnumProc(IntPtr hMonitor, IntPtr hdcMonitor, ref RECT lprcMonitor, IntPtr dwData)
        {
            if (!GetNumberOfPhysicalMonitorsFromHMONITOR(hMonitor, out uint count) || count == 0)
                return true;

            var physicalMonitors = new PHYSICAL_MONITOR[count];
            if (!GetPhysicalMonitorsFromHMONITOR(hMonitor, count, physicalMonitors))
                return true;

            foreach (var pm in physicalMonitors)
            {
                // Probe DDC/CI support; monitors (or cables) without it fail here.
                if (GetVCPFeatureAndVCPFeatureReply(pm.hPhysicalMonitor, VcpBrightness, IntPtr.Zero, out _, out uint max))
                {
                    index++;
                    var label = string.IsNullOrWhiteSpace(pm.szPhysicalMonitorDescription)
                        ? $"외부 모니터 {index}"
                        : $"외부 모니터 {index} ({pm.szPhysicalMonitorDescription})";
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

    [StructLayout(LayoutKind.Sequential)]
    private struct RECT
    {
        public int Left, Top, Right, Bottom;
    }

    [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Unicode)]
    private struct PHYSICAL_MONITOR
    {
        public IntPtr hPhysicalMonitor;
        [MarshalAs(UnmanagedType.ByValTStr, SizeConst = 128)]
        public string szPhysicalMonitorDescription;
    }

    private delegate bool MonitorEnumProc(IntPtr hMonitor, IntPtr hdcMonitor, ref RECT lprcMonitor, IntPtr dwData);

    [DllImport("user32.dll")]
    private static extern bool EnumDisplayMonitors(IntPtr hdc, IntPtr lprcClip, MonitorEnumProc lpfnEnum, IntPtr dwData);

    [DllImport("dxva2.dll")]
    private static extern bool GetNumberOfPhysicalMonitorsFromHMONITOR(IntPtr hMonitor, out uint pdwNumberOfPhysicalMonitors);

    [DllImport("dxva2.dll", CharSet = CharSet.Unicode)]
    private static extern bool GetPhysicalMonitorsFromHMONITOR(IntPtr hMonitor, uint dwPhysicalMonitorArraySize,
        [Out] PHYSICAL_MONITOR[] pPhysicalMonitorArray);

    [DllImport("dxva2.dll")]
    private static extern bool DestroyPhysicalMonitors(uint dwPhysicalMonitorArraySize,
        [In] PHYSICAL_MONITOR[] pPhysicalMonitorArray);

    [DllImport("dxva2.dll")]
    private static extern bool GetVCPFeatureAndVCPFeatureReply(IntPtr hMonitor, byte bVCPCode, IntPtr pvct,
        out uint pdwCurrentValue, out uint pdwMaximumValue);

    [DllImport("dxva2.dll")]
    private static extern bool SetVCPFeature(IntPtr hMonitor, byte bVCPCode, uint dwNewValue);
}
