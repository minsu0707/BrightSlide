using System.Management;

namespace BrightSlide.Services;

/// <summary>
/// Controls the brightness of a built-in panel (laptop screen) via the
/// root\WMI WmiMonitorBrightness / WmiMonitorBrightnessMethods classes.
/// </summary>
public sealed class InternalDisplayController : IDisplayController
{
    private readonly string _instanceName;

    public string Id => _instanceName;
    public string Name { get; }

    public InternalDisplayController(string instanceName, string name)
    {
        _instanceName = instanceName;
        Name = name;
    }

    public static IReadOnlyList<InternalDisplayController> DiscoverAll()
    {
        var result = new List<InternalDisplayController>();
        try
        {
            using var searcher = new ManagementObjectSearcher("root\\WMI", "SELECT * FROM WmiMonitorBrightness");
            using var collection = searcher.Get();
            int index = 0;
            foreach (ManagementBaseObject instance in collection)
            {
                var instanceName = (string)instance["InstanceName"];
                var label = collection.Count > 1 ? $"내장 디스플레이 {++index}" : "내장 디스플레이";
                result.Add(new InternalDisplayController(instanceName, label));
                instance.Dispose();
            }
        }
        catch (ManagementException)
        {
            // WMI brightness class not present on this machine (common on desktops).
        }
        return result;
    }

    public int GetBrightness()
    {
        using var searcher = new ManagementObjectSearcher("root\\WMI",
            $"SELECT * FROM WmiMonitorBrightness WHERE InstanceName = '{Escape(_instanceName)}'");
        foreach (ManagementBaseObject instance in searcher.Get())
        {
            using (instance)
            {
                return Convert.ToInt32(instance["CurrentBrightness"]);
            }
        }
        return 0;
    }

    public void SetBrightness(int percent)
    {
        percent = Math.Clamp(percent, 0, 100);
        using var searcher = new ManagementObjectSearcher("root\\WMI",
            $"SELECT * FROM WmiMonitorBrightnessMethods WHERE InstanceName = '{Escape(_instanceName)}'");
        foreach (ManagementObject instance in searcher.Get())
        {
            using (instance)
            {
                var inParams = instance.GetMethodParameters("WmiSetBrightness");
                inParams["Timeout"] = 0;
                inParams["Brightness"] = (byte)percent;
                instance.InvokeMethod("WmiSetBrightness", inParams, null);
            }
            return;
        }
    }

    private static string Escape(string value) => value.Replace("'", "''").Replace("\\", "\\\\");
}
