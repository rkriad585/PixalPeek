; PixalPeek Windows Installer - NSIS Script
; Requires: NSIS (https://nsis.sourceforge.io)

!include "MUI2.nsh"
!include "FileFunc.nsh"

; ── Configuration ──────────────────────────────────────
!ifndef APP_NAME
  !define APP_NAME        "PixalPeek"
!endif
!ifndef APP_VERSION
  !define APP_VERSION     "0.1.5-beta"
!endif
!ifndef APP_ARCH
  !define APP_ARCH        "amd64"
!endif
!define APP_PUBLISHER   "rkriad585"
!define APP_URL         "https://rkriad585.github.io/PixalPeek"
!define APP_EXE         "pixalpeek.exe"
!define APP_DESC        "Dot-matrix QR code scanner & generator"

!define INSTALL_DIR     "$LOCALAPPDATA\${APP_NAME}"
!define UNINSTALL_KEY   "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"

Name "${APP_NAME} ${APP_VERSION}"
OutFile "pixalpeek-windows-${APP_ARCH}.exe"
InstallDir "${INSTALL_DIR}"
InstallDirRegKey HKCU "${INSTALL_DIR}" ""
RequestExecutionLevel user

; ── Interface ──────────────────────────────────────────
!define MUI_ICON              "build\appicon.ico"
!define MUI_UNICON            "build\appicon.ico"
!define MUI_ABORTWARNING
!define MUI_WELCOMEPAGE_TITLE "Welcome to ${APP_NAME} Setup"
!define MUI_WELCOMEPAGE_TEXT  "This will install ${APP_NAME} ${APP_VERSION}.$\r$\n$\r$\n${APP_DESC}"

; ── Pages ──────────────────────────────────────────────
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "..\..\LICENSE"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

; ── Installer ──────────────────────────────────────────
Section "Install" SecInstall
    SetOutPath "$INSTDIR"
    SetShellVarContext current

    File "..\..\pixalpeek.exe"

    ; Write uninstaller
    WriteUninstaller "$INSTDIR\uninstall.exe"

    ; Start Menu shortcuts
    CreateDirectory "$SMPROGRAMS\${APP_NAME}"
    CreateShortcut "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}"
    CreateShortcut "$SMPROGRAMS\${APP_NAME}\Uninstall.lnk" "$INSTDIR\uninstall.exe"

    ; Registry (for uninstaller)
    WriteRegStr HKCU "${UNINSTALL_KEY}" "DisplayName" "${APP_NAME}"
    WriteRegStr HKCU "${UNINSTALL_KEY}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr HKCU "${UNINSTALL_KEY}" "DisplayVersion" "${APP_VERSION}"
    WriteRegStr HKCU "${UNINSTALL_KEY}" "Publisher" "${APP_PUBLISHER}"
    WriteRegStr HKCU "${UNINSTALL_KEY}" "URLInfoAbout" "${APP_URL}"
    WriteRegStr HKCU "${UNINSTALL_KEY}" "InstallLocation" "$INSTDIR"
    WriteRegDWORD HKCU "${UNINSTALL_KEY}" "NoModify" 1
    WriteRegDWORD HKCU "${UNINSTALL_KEY}" "NoRepair" 1

    ; Add to PATH
    EnVar::AddValue "PATH" "$INSTDIR"

    ; File size
    ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
    IntFmt $0 "0x%08X" $0
    WriteRegDWORD HKCU "${UNINSTALL_KEY}" "EstimatedSize" "$0"
SectionEnd

; ── Uninstaller ────────────────────────────────────────
Section "Uninstall"
    SetShellVarContext current

    Delete "$INSTDIR\${APP_EXE}"
    Delete "$INSTDIR\uninstall.exe"
    RMDir "$INSTDIR"

    RMDir /r "$SMPROGRAMS\${APP_NAME}"

    DeleteRegKey HKCU "${UNINSTALL_KEY}"

    EnVar::RemoveValue "PATH" "$INSTDIR"
SectionEnd
