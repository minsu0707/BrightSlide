#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif
#ifndef StageDir
  #define StageDir "stage"
#endif

#define AppName "BrightSlide"
#define AppExe "BrightSlide.exe"

[Setup]
AppId={{5B7E2C9A-3D41-4F8E-9A62-7C1D0E4B8F35}
AppName={#AppName}
AppVersion={#AppVersion}
AppVerName={#AppName} {#AppVersion}
AppPublisher=minsu0707
AppPublisherURL=https://github.com/minsu0707/BrightSlide
AppSupportURL=https://github.com/minsu0707/BrightSlide/issues
AppUpdatesURL=https://github.com/minsu0707/BrightSlide/releases
VersionInfoVersion={#AppVersion}
DefaultDirName={autopf}\{#AppName}
DefaultGroupName={#AppName}
DisableProgramGroupPage=yes
PrivilegesRequired=lowest
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
MinVersion=10.0
OutputDir=Output
OutputBaseFilename=BrightSlide-Setup-{#AppVersion}
SetupIconFile=..\Assets\brightslide.ico
UninstallDisplayIcon={app}\{#AppExe}
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern

[Languages]
Name: "korean"; MessagesFile: "compiler:Languages\Korean.isl"

[Tasks]
Name: "startup"; Description: "Windows 시작 시 자동 실행"; Flags: unchecked
Name: "desktopicon"; Description: "바탕 화면에 바로 가기 만들기"; Flags: unchecked

[Files]
Source: "{#StageDir}\{#AppExe}"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{autoprograms}\{#AppName}"; Filename: "{app}\{#AppExe}"
Name: "{autodesktop}\{#AppName}"; Filename: "{app}\{#AppExe}"; Tasks: desktopicon

[Registry]
; Unquoted on purpose: StartupManager.IsEnabled compares this value to the exe path verbatim.
Root: HKCU; Subkey: "Software\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; ValueName: "{#AppName}"; ValueData: "{app}\{#AppExe}"; Tasks: startup

[Run]
Filename: "{app}\{#AppExe}"; Description: "{#AppName} 지금 실행"; Flags: nowait postinstall skipifsilent

[Code]
procedure StopRunningApp;
var
  ResultCode: Integer;
begin
  Exec(ExpandConstant('{sys}\taskkill.exe'), '/F /IM {#AppExe}', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
end;

function PrepareToInstall(var NeedsRestart: Boolean): String;
begin
  StopRunningApp;
  Result := '';
end;

function InitializeUninstall: Boolean;
begin
  StopRunningApp;
  Result := True;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  // The in-app "시작 시 자동 실행" toggle writes this value too, so remove it regardless of how it got there.
  if CurUninstallStep = usPostUninstall then
    RegDeleteValue(HKCU, 'Software\Microsoft\Windows\CurrentVersion\Run', '{#AppName}');
end;
