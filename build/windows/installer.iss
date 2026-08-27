; PixalPeek - Inno Setup script
; Compile: ISCC /DAPP_VERSION=0.1.5-beta /DAPP_ARCH=amd64 installer.iss

#ifndef APP_VERSION
  #define APP_VERSION "0.1.5-beta"
#endif
#ifndef APP_ARCH
  #define APP_ARCH "amd64"
#endif

#define APP_NAME "PixalPeek"
#define APP_EXE "pixalpeek.exe"
#define APP_PUBLISHER "rkriad585"
#define APP_URL "https://rkriad585.github.io/PixalPeek"
#define APP_DESC "QR code scanner & generator"

[Setup]
AppId={{B11A4C3E-6A7E-4C2D-9F13-9D8B7E2C4A11}
AppName={#APP_NAME}
AppVersion={#APP_VERSION}
AppPublisher={#APP_PUBLISHER}
AppPublisherURL={#APP_URL}
AppSupportURL={#APP_URL}
AppUpdatesURL={#APP_URL}
DefaultDirName={localappdata}\{#APP_NAME}
DefaultGroupName={#APP_NAME}
DisableProgramGroupPage=yes
UninstallDisplayIcon={app}\appicon.ico
UninstallDisplayName={#APP_NAME}
OutputDir={#SourcePath}..\..\dist
OutputBaseFilename=pixalpeek-windows-{#APP_ARCH}
SetupIconFile={#SourcePath}..\appicon.ico
Compression=lzma2/ultra
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=lowest
ArchitecturesInstallIn64BitMode=x64compatible
ArchitecturesAllowed=x64compatible
ChangesEnvironment=yes
VersionInfoVersion=0.1.5.0

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
Source: "{#SourcePath}..\..\dist\pixalpeek.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#SourcePath}..\appicon.ico"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{autoprograms}\{#APP_NAME}\{#APP_NAME}"; Filename: "{app}\{#APP_EXE}"; IconFilename: "{app}\appicon.ico"; WorkingDir: "{app}"
Name: "{autoprograms}\{#APP_NAME}\Uninstall {#APP_NAME}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#APP_NAME}"; Filename: "{app}\{#APP_EXE}"; IconFilename: "{app}\appicon.ico"; WorkingDir: "{app}"; Tasks: desktopicon

[Registry]
Root: HKCU; Subkey: "Software\Microsoft\Windows\CurrentVersion\Uninstall\{#APP_NAME}"; ValueType: string; ValueName: "DisplayIcon"; ValueData: "{app}\appicon.ico"; Flags: uninsdeletevalue

[Run]
Filename: "{app}\{#APP_EXE}"; Description: "{cm:LaunchProgram,{#APP_NAME}}"; Flags: nowait postinstall skipifsilent

[Code]
const
  EnvKey = 'Environment';
  PathValueName = 'Path';

// Add or remove {app} from the user PATH (HKCU\Environment) at the registry level.
// ChangesEnvironment=yes asks Inno to broadcast the change automatically.

function GetUserPath(): string;
begin
  Result := '';
  if RegQueryStringValue(HKCU, EnvKey, PathValueName, Result) then
  begin
    // keep as-is
  end;
end;

function IsInPath(const PathToAdd: string; const Path: string): Boolean;
var
  lowerPath, lowerP, token: string;
  i: Integer;
begin
  Result := False;
  lowerPath := LowerCase(Path);
  lowerP := LowerCase(PathToAdd);
  // Token by token on ';'
  i := 1;
  while i <= Length(lowerPath) do
  begin
    token := '';
    while (i <= Length(lowerPath)) and (lowerPath[i] <> ';') do
    begin
      token := token + lowerPath[i];
      i := i + 1;
    end;
    if token = lowerP then
    begin
      Result := True;
      Exit;
    end;
    i := i + 1; // skip ';'
  end;
end;

procedure RemoveFromPath(const PathToRemove: string; var Path: string);
var
  lowerPath: string;
  token: string;
  parts, remaining: string;
  i: Integer;
begin
  lowerPath := LowerCase(PathToRemove);
  remaining := Path;
  parts := '';
  i := 1;
  while i <= Length(remaining) do
  begin
    token := '';
    while (i <= Length(remaining)) and (remaining[i] <> ';') do
    begin
      token := token + remaining[i];
      i := i + 1;
    end;
    if LowerCase(token) <> lowerPath then
    begin
      if parts <> '' then parts := parts + ';';
      parts := parts + token;
    end;
    if i <= Length(remaining) then i := i + 1; // skip ';'
  end;
  Path := parts;
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  appPath, curPath, newPath: string;
begin
  if CurStep = ssPostInstall then
  begin
    appPath := ExpandConstant('{app}');
    curPath := GetUserPath();
    if not IsInPath(appPath, curPath) then
    begin
      if curPath = '' then
        newPath := appPath
      else
        newPath := curPath + ';' + appPath;
      RegWriteExpandStringValue(HKCU, EnvKey, PathValueName, newPath);
    end;
  end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
  appPath, curPath, newPath: string;
begin
  if CurUninstallStep = usPostUninstall then
  begin
    appPath := ExpandConstant('{app}');
    curPath := GetUserPath();
    if curPath <> '' then
    begin
      newPath := curPath;
      RemoveFromPath(appPath, newPath);
      if newPath = '' then
        RegDeleteValue(HKCU, EnvKey, PathValueName)
      else
        RegWriteExpandStringValue(HKCU, EnvKey, PathValueName, newPath);
    end;
  end;
end;
