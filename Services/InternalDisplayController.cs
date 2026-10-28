using System.Management;

namespace BrightSlide.Services;

/// <summary>
/// Controls the brightness of a built-in panel (laptop screen) via the
/// root\WMI WmiMonitorBrightness / WmiMonitorBrightnessMethods classes.
/// </summary>
public sealed class InternalDisplayController : IDisplayController
{
    private readonly string _instanceName;

    // Cached WMI handles: bound once to the instance's ManagementPath so every
    // Get/Set reuses the same object instead of re-running a SELECT query
    // (which otherwise adds tens of ms of overhead per drag tick).
    private readonly ManagementObject _brightnessObject;
    private readonly ManagementObject _methodsObject;

    public string Id => _instanceName;
    public string Name { get; }

    public InternalDisplayController(string instanceName, string name)
    {
        _instanceName = instanceName;
        Name = name;

        var escaped = Escape(instanceName);
        _brightnessObject = new ManagementObject("root\\WMI",
            $"WmiMonitorBrightness.InstanceName='{escaped}'", null);
        _methodsObject = new ManagementObject("root\\WMI",
            $"WmiMonitorBrightnessMethods.InstanceName='{escaped}'", null);
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
        _brightnessObject.Get();
        return Convert.ToInt32(_brightnessObject["CurrentBrightness"]);
    }

    public void SetBrightness(int percent)
    {
        percent = Math.Clamp(percent, 0, 100);
        var inParams = _methodsObject.GetMethodParameters("WmiSetBrightness");
        inParams["Timeout"] = 0;
        inParams["Brightness"] = (byte)percent;
        _methodsObject.InvokeMethod("WmiSetBrightness", inParams, null);
    }

    private static string Escape(string value) => value.Replace("'", "''").Replace("\\", "\\\\");
}
