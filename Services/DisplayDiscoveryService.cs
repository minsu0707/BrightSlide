using BrightSlide.Models;

namespace BrightSlide.Services;

public sealed class DisplayDiscoveryService
{
    private List<ExternalDisplayController> _externalControllers = new();

    public List<DisplayDevice> DiscoverAll()
    {
        var devices = new List<DisplayDevice>();

        devices.AddRange(InternalDisplayController.DiscoverAll()
            .Select(controller => CreateDevice(controller, DisplayKind.Internal)));

        _externalControllers = ExternalDisplayController.DiscoverAll();
        devices.AddRange(_externalControllers
            .Select(controller => CreateDevice(controller, DisplayKind.External)));

        return devices;
    }

    public void ReleaseExternalHandles() => ExternalDisplayController.ReleaseAll(_externalControllers);

    private static DisplayDevice CreateDevice(IDisplayController controller, DisplayKind kind) => new()
    {
        Id = controller.Id,
        Name = controller.Name,
        Kind = kind,
        Controller = controller,
        Brightness = controller.GetBrightness()
    };
}
