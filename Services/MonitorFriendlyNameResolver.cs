using System.Management;

namespace BrightSlide.Services;

/// <summary>
/// Reads real monitor model names from EDID data (root\WMI WmiMonitorID),
/// keyed by the hardware-id segment ("SKG2712") shared with the DeviceID
/// string returned by EnumDisplayDevices, so a physical monitor handle can be
/// matched to its human-readable model name instead of Windows' generic
/// "Generic PnP Monitor" description.
/// </summary>
internal static class MonitorFriendlyNameResolver
{
    public static Dictionary<string, string> LoadByHardwareId()
    {
        var map = new Dictionary<string, string>(StringComparer.OrdinalIgnoreCase);
        try
        {
            using var searcher = new ManagementObjectSearcher("root\\wmi", "SELECT * FROM WmiMonitorID");
            foreach (ManagementBaseObject instance in searcher.Get())
            {
                using (instance)
                {
                    var instanceName = (string)instance["InstanceName"];
                    var manufacturer = DecodeText(instance["ManufacturerName"] as ushort[]);
                    var product = DecodeText(instance["ProductCodeID"] as ushort[]);
                    var friendly = DecodeText(instance["UserFriendlyName"] as ushort[]);

                    var name = !string.IsNullOrWhiteSpace(friendly)
                        ? friendly
                        : $"{manufacturer} {product}".Trim();

                    // InstanceName looks like "DISPLAY\SKG2712\5&32428203&0&UID4353_0";
                    // the second segment is the same hardware-id EnumDisplayDevices exposes.
                    var segments = instanceName.Split('\\');
                    if (segments.Length >= 2 && !string.IsNullOrWhiteSpace(name))
                        map[segments[1]] = name;
                }
            }
        }
        catch (ManagementException)
        {
            // No monitor EDID info available; callers fall back to the DDC/CI description.
        }
        return map;
    }

    private static string DecodeText(ushort[]? values)
    {
        if (values is null) return string.Empty;
        return new string(values.Where(v => v != 0).Select(v => (char)v).ToArray());
    }
}
