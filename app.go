package main

import (
	"context"
	_ "embed"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
	"encoding/json"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"regexp" // Added this import
)

type WailsJson struct {
	Name string `json:"name"`
}

//go:embed executables/OTPGenerator.exe
var otpGeneratorExe []byte

//go:embed executables/ZEPSdkInvokeOTP.exe
var zepSdkInvokeOtpExe []byte

//go:embed executables/ZDPObfuscate.exe
var zdpObfuscateExe []byte

//go:embed resources/dlp_config_dlp_sdk.json
var dlpConfigDlpSdkJson []byte

const (
	defaultKeyFilePath = `C:\ProgramData\Zscaler\ZDP\Settings\zdp_endpoint_id`
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// --- Methods callable from frontend ---

func (a *App) IsZdpServiceRunning() bool {
	cmd := exec.Command("powershell", "-Command", "Get-Service -Name zdpservice")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "Running")
}

func (a *App) GetDetailsHttpsCmd() (string, error) {
	url := "https://127.0.0.1:9861/api/v1.0/get-zdpe-details"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTPS request: %w", err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTPS request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read HTTPS response body: %w", err)
	}
	return string(body), nil
}

func (a *App) GetDetailsHttpCmd() (string, error) {
	url := "http://127.0.0.1:9861/api/v1.0/get-zdpe-details"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read HTTP response body: %w", err)
	}
	return string(body), nil
}

func (a *App) EnableAntiTampering() error {
	return a.setAntiTamperingMode(true)
}

func (a *App) DisableAntiTampering() error {
	return a.setAntiTamperingMode(false)
}

func (a *App) GetAntiTamperingStatus() (string, error) {
	output, err := a.runEmbeddedExe(zepSdkInvokeOtpExe, "ZEPSdkInvokeOTP.exe", "GetATMode")
	if err != nil {
		return "", err
	}
	if strings.Contains(output, "Enabled") {
		return "Enabled", nil
	}
	if strings.Contains(output, "Disabled") {
		return "Disabled", nil
	}
	return "Unknown", nil
}

func (a *App) DeobfuscateOotbSettings() (string, error) {
	obfuscated, err := a.IsOotbSettingsObfuscated()
	if err != nil {
		return "", err
	}
	if !obfuscated {
		return "File is already de-obfuscated.", nil
	}

	filePath := `C:\ProgramData\Zscaler\ZDP\Settings\zdp_endpoint_settings_ootb.json`
	err = a.runZDPObfuscate(true, defaultKeyFilePath, filePath)
	if err != nil {
		if strings.Contains(err.Error(), "key file read error") {
			return "", fmt.Errorf("key file read error: Anti-tampering is enabled. Please disable it to de-obfuscate the file")
		}
		return "", fmt.Errorf("failed to de-obfuscate ootb-settings: %w", err)
	}
	return fmt.Sprintf("File '%s' de-obfuscated successfully.", filePath), nil
}

func (a *App) DeobfuscateZdpModes() (string, error) {
	obfuscated, err := a.IsZdpModesObfuscated()
	if err != nil {
		return "", err
	}
	if !obfuscated {
		return "File is already de-obfuscated.", nil
	}

	filePath := `C:\ProgramData\Zscaler\ZDP\Settings\zdp_modes.json`
	err = a.runZDPObfuscate(true, defaultKeyFilePath, filePath)
	if err != nil {
		if strings.Contains(err.Error(), "key file read error") {
			return "", fmt.Errorf("key file read error: Anti-tampering is enabled. Please disable it to de-obfuscate the file")
		}
		return "", fmt.Errorf("failed to de-obfuscate zdp-modes: %w", err)
	}
	return fmt.Sprintf("File '%s' de-obfuscated successfully.", filePath), nil
}

// --- Helper Functions ---

func (a *App) runZDPObfuscate(deobfuscate bool, keyPath, filePath string) error {
	var args []string
	if deobfuscate {
		args = append(args, "-d")
	}
	args = append(args, keyPath, filePath)
	_, err := a.runEmbeddedExe(zdpObfuscateExe, "ZDPObfuscate.exe", args...)
	return err
}

func (a *App) setAntiTamperingMode(enable bool) error {
	otp, err := a.getOTP()
	if err != nil {
		return err
	}
	mode := "0"
	if enable {
		mode = "1"
	}
	_, err = a.runEmbeddedExe(zepSdkInvokeOtpExe, "ZEPSdkInvokeOTP.exe", "SetATModeEx", mode, otp)
	return err
}

func (a *App) getOTP() (string, error) {
	hostname, err := a.getHostname()
	if err != nil {
		return "", err
	}
	output, err := a.runEmbeddedExe(otpGeneratorExe, "OTPGenerator.exe", hostname)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}
	otp := strings.TrimSpace(strings.TrimPrefix(output, "OTP:"))
	if otp == "" {
		return "", fmt.Errorf("OTPGenerator.exe returned empty OTP")
	}
	return otp, nil
}

func (a *App) getHostname() (string, error) {
	return os.Hostname()
}

func (a *App) runEmbeddedExe(exeData []byte, exeName string, args ...string) (string, error) {
	tempDir, err := ioutil.TempDir("", "go-app-exec")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	tempExePath := filepath.Join(tempDir, exeName)
	err = ioutil.WriteFile(tempExePath, exeData, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to write embedded executable to temp file: %w", err)
	}

	cmd := exec.Command(tempExePath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("executable '%s' failed: %w\nOutput: %s", exeName, err, string(outputBytes))
	}
	return string(outputBytes), nil
}

func (a *App) GetVersion() (string, error) {
	wailsJsonFile, err := os.ReadFile("wails.json")
	if err != nil {
		return "", fmt.Errorf("failed to read wails.json: %w", err)
	}

	var wailsJson WailsJson
	err = json.Unmarshal(wailsJsonFile, &wailsJson)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal wails.json: %w", err)
	}

	return wailsJson.Name, nil
}

func (a *App) IsOotbSettingsObfuscated() (bool, error) {
	return a.isFileObfuscated(`C:\ProgramData\Zscaler\ZDP\Settings\zdp_endpoint_settings_ootb.json`)
}

func (a *App) IsZdpModesObfuscated() (bool, error) {
	return a.isFileObfuscated(`C:\ProgramData\Zscaler\ZDP\Settings\zdp_modes.json`)
}

func (a *App) isFileObfuscated(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 4)
	_, err = file.Read(buffer)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	return string(buffer) == "ZDPU", nil
}

type ClassifierOutput struct {
	Command           string `json:"command"`
	Output            string `json:"output"`
	OcrTextPath       string `json:"ocrTextPath"`
	ExtractedTextPath string `json:"extractedTextPath"`
}

var (
	versionDLL               = syscall.NewLazyDLL("version.dll")
	procGetFileVersionInfoSize = versionDLL.NewProc("GetFileVersionInfoSizeW")
	procGetFileVersionInfo     = versionDLL.NewProc("GetFileVersionInfoW")
	procVerQueryValue          = versionDLL.NewProc("VerQueryValueW")
)

type VS_FIXEDFILEINFO struct {
	dwSignature        uint32
	dwStrucVersion     uint32
	dwFileVersionMS    uint32
	dwFileVersionLS    uint32
	dwProductVersionMS uint32
	dwProductVersionLS uint32
	dwFileFlagsMask    uint32
	dwFileFlags        uint32
	dwFileOS           uint32
	dwFileType         uint32
	dwFileSubtype      uint32
	dwFileDateMS       uint32
	dwFileDateLS       uint32
}

type AllVersions struct {
    Zdp string `json:"zdp"`
    Zcc string `json:"zcc"`
    Zep string `json:"zep"`
}

func (a *App) getExeVersion(filePath string) (string, error) {
	filePathPtr, err := syscall.UTF16PtrFromString(filePath)
	if err != nil {
		return "", err
	}

	infoSize, _, err := procGetFileVersionInfoSize.Call(uintptr(unsafe.Pointer(filePathPtr)), 0)
	if infoSize == 0 {
		return "", fmt.Errorf("GetFileVersionInfoSizeW failed: %v", err)
	}

	infoBuf := make([]byte, infoSize)

	ret, _, err := procGetFileVersionInfo.Call(
		uintptr(unsafe.Pointer(filePathPtr)),
		0,
		uintptr(infoSize),
		uintptr(unsafe.Pointer(&infoBuf[0])),
	)
	if ret == 0 {
		return "", fmt.Errorf("GetFileVersionInfoW failed: %v", err)
	}

	var fixedInfo *VS_FIXEDFILEINFO
	var len uint32
	ret, _, err = procVerQueryValue.Call(
		uintptr(unsafe.Pointer(&infoBuf[0])),
		uintptr(unsafe.Pointer(syscall.StringBytePtr(`\`))),
		uintptr(unsafe.Pointer(&fixedInfo)),
		uintptr(unsafe.Pointer(&len)),
	)
	if ret == 0 {
		return "", fmt.Errorf("VerQueryValueW failed: %v", err)
	}

	verMS := fixedInfo.dwProductVersionMS
	verLS := fixedInfo.dwProductVersionLS
	major := (verMS >> 16) & 0xffff
	minor := verMS & 0xffff
	patch := (verLS >> 16) & 0xffff
	build := verLS & 0xffff
    
	return fmt.Sprintf("%d.%d.%d.%d", major, minor, patch, build), nil
}


func (a *App) GetAllVersions() (*AllVersions, error) {
    zdpPath := `C:\Program Files\Zscaler\ZDP\ZDPService.exe`
    zccPath := `C:\Program Files\Zscaler\ZSATray\ZSATray.exe`
    zepPath := `C:\Program Files\Zscaler\ZEP\ZEPService.exe`

    zdpVer, zdpErr := a.getExeVersion(zdpPath)
    if zdpErr != nil {
        zdpVer = "Not Found"
    }

    zccVer, zccErr := a.getExeVersion(zccPath)
    if zccErr != nil {
        zccVer = "Not Found"
    }
    
    zepVer, zepErr := a.getExeVersion(zepPath)
    if zepErr != nil {
        zepVer = "Not Found"
    }

    return &AllVersions{
        Zdp: zdpVer,
        Zcc: zccVer,
        Zep: zepVer,
    }, nil
}


func (a *App) StandaloneClassifier(filePath string, configOption string, configPath string, useOcr bool, useText bool) (*ClassifierOutput, error) {
	classifierPath := `C:\Program Files\Zscaler\ZDP\ZDPClassifier.exe`
	if _, err := os.Stat(classifierPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("classifier executable not found at %s", classifierPath)
	}

	var finalConfigPath string
	switch configOption {
	case "default":
		tempDir, err := ioutil.TempDir("", "zdp-tool-")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp dir: %w", err)
		}
		defer os.RemoveAll(tempDir)

		finalConfigPath = filepath.Join(tempDir, "dlp_config_dlp_sdk.json")
		err = ioutil.WriteFile(finalConfigPath, dlpConfigDlpSdkJson, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to write embedded config to temp file: %w", err)
		}
	case "last_modified":
		configDir := `C:\ProgramData\Zscaler\ZDP\Config`
		latestConfig, err := a.getLatestConfigFile(configDir)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest config file: %w", err)
		}
		finalConfigPath = latestConfig
	case "custom":
		finalConfigPath = configPath
	default:
		return nil, fmt.Errorf("invalid config option: %s", configOption)
	}

	var args []string
	args = append(args, "-config", finalConfigPath, "-file", filePath)
	if useOcr {
		args = append(args, "-ocr")
	}
	if useText {
		args = append(args, "-text")
	}

	cmdString := fmt.Sprintf("%s %s", classifierPath, strings.Join(args, " "))
	cmd := exec.Command(classifierPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run classifier: %w\nOutput: %s", err, string(output))
	}

	var ocrTextPath, extractedTextPath string
	if useOcr {
		ocrTextPath = filePath + ".ocr_text.txt"
	}
	if useText {
		extractedTextPath = filePath + ".extracted_text.txt"
	}

	return &ClassifierOutput{
		Command:           cmdString,
		Output:            string(output),
		OcrTextPath:       ocrTextPath,
		ExtractedTextPath: extractedTextPath,
	}, nil
}

func (a *App) getLatestConfigFile(dir string) (string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var latestFile os.FileInfo
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			if latestFile == nil || file.ModTime().After(latestFile.ModTime()) {
				latestFile = file
			}
		}
	}

	if latestFile == nil {
		return "", fmt.Errorf("no json files found in %s", dir)
	}

	return filepath.Join(dir, latestFile.Name()), nil
}

func (a *App) SelectFile() (string, error) {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select File",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", err
	}
	return file, nil
}

func (a *App) ReadFileContent(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return string(content), nil
}
// GetDlpSdkVersion function (added manually because I kept messing up the replace)
func (a *App) GetDlpSdkVersion() (string, error) {
	fmt.Println("Attempting to get DLP SDK Version...")
	result, classifierErr := a.StandaloneClassifier("C:\\ProgramData\\Zscaler\\ZDP\\Logs\\zdp_install.log", "default", "", false, false)

	// Check if result is nil first
	if result == nil {
		if classifierErr != nil {
			fmt.Printf("StandaloneClassifier returned nil result with error: %v\n", classifierErr)
			return "Unknown", fmt.Errorf("StandaloneClassifier returned nil result: %w", classifierErr)
		}
		// This case should ideally not happen (nil result without an error)
		fmt.Println("StandaloneClassifier returned nil result without an error. This is unexpected.")
		return "Unknown", fmt.Errorf("StandaloneClassifier returned nil result unexpectedly")
	}
	
	// Proceed with logging and parsing only if result is not nil
	fmt.Printf("Executed Command: %s\n", result.Command)
	fmt.Printf("Classifier Output: %s\n", result.Output)
	if classifierErr != nil {
		fmt.Printf("StandaloneClassifier returned an error (ignored if version found): %v\n", classifierErr)
	}

	// Always attempt to parse the DLP SDK version from the output
	re := regexp.MustCompile(`DLP SDK version: (.*)`)
	match := re.FindStringSubmatch(result.Output)
	
	fmt.Printf("Regex Match Result: %v\n", match)

	if len(match) > 1 {
		version := strings.TrimSpace(match[1])
		fmt.Printf("DLP SDK Version Extracted: %s\n", version)
		return version, nil
	}
	fmt.Println("DLP SDK Version not found in output.")

	// If version is not found and classifier had an error, then return the error
	if classifierErr != nil {
		return "Unknown", fmt.Errorf("DLP SDK version not found in output and classifier failed: %w", classifierErr)
	}

	return "Unknown", nil
}