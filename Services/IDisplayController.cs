namespace BrightSlide.Services;

public interface IDisplayController
{
    string Id { get; }
    string Name { get; }
    int GetBrightness();
    void SetBrightness(int percent);
}
