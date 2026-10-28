using System.Runtime.InteropServices;

namespace BrightSlide.Interop;

internal static class CursorNativeMethods
{
    [StructLayout(LayoutKind.Sequential)]
    public struct POINT
    {
        public int X;
        public int Y;
    }

    [DllImport("user32.dll")]
    public static extern bool GetCursorPos(out POINT lpPoint);
}
