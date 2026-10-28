using BrightSlide.Models;

namespace BrightSlide.Services;

public sealed class DisplayDiscoveryService
{
    private List<ExternalDisplayController> _externalControllers = new();

    public List<DisplayDevice> DiscoverAll()
    {
        var devices = new List<DisplayDevice>();

        foreach (var internalController in InternalDisplayController.DiscoverAll())
        {
            devices.Add(new DisplayDevice
            {
                Id = internalController.Id,
                Name = internalController.Name,
                Kind = DisplayKind.Internal,
                Controller = internalController,
                Brightness = internalController.GetBrightness()
            });
        }

        _externalControllers = ExternalDisplayController.DiscoverAll();
        foreach (var externalController in _externalControllers)
        {
            devices.Add(new DisplayDevice
            {
                Id = externalController.Id,
                Name = externalController.Name,
                Kind = DisplayKind.External,
                Controller = externalController,
                Brightness = externalController.GetBrightness()
            });
        }

        return devices;
    }

    public void ReleaseExternalHandles() => ExternalDisplayController.ReleaseAll(_externalControllers);
}
