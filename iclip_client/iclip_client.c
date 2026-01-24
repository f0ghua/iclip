/*
 * iclip_client - Clipboard transfer client for Linux/Ubuntu
 *
 * Monitors the system clipboard and transfers text content to a Windows PC
 * via SMB share. Uses only libc and system tools (xclip, tar, smbclient).
 *
 * Compile: gcc -o iclip_client iclip_client.c -O2
 * Or use: make
 */

#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <time.h>
#include <signal.h>
#include <termios.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/select.h>
#include <pwd.h>
#include <errno.h>

#define VERSION "1.0"
#define MAX_CLIPBOARD_SIZE (1024 * 1024)  /* 1MB max clipboard content */
#define MAX_PATH_LEN 1024
#define MAX_CMD_LEN 2048
#define MAX_LINE_LEN 256
#define CONFIG_FILE "config.ini"

/* Configuration structure */
typedef struct {
    char server[128];
    char share[64];
    char path[128];
    char username[64];
    char password[128];
    char workgroup[64];
    float interval;
} Config;

/* UI State */
typedef struct {
    char status[32];
    char last_sync[16];
    char clipboard_preview[256];
    char error_msg[128];
    int paused;
} UIState;

/* Global variables */
static volatile sig_atomic_t g_running = 1;
static struct termios g_orig_termios;
static int g_term_modified = 0;

/* Signal handler */
static void signal_handler(int sig) {
    (void)sig;
    g_running = 0;
}

/* Restore terminal settings */
static void restore_terminal(void) {
    if (g_term_modified) {
        tcsetattr(STDIN_FILENO, TCSAFLUSH, &g_orig_termios);
        g_term_modified = 0;
    }
}

/* Set terminal to raw mode for non-blocking key input */
static void set_raw_terminal(void) {
    struct termios raw;

    if (!isatty(STDIN_FILENO)) return;

    if (tcgetattr(STDIN_FILENO, &g_orig_termios) == -1) return;

    raw = g_orig_termios;
    raw.c_lflag &= ~(ICANON | ECHO);
    raw.c_cc[VMIN] = 0;
    raw.c_cc[VTIME] = 0;

    if (tcsetattr(STDIN_FILENO, TCSAFLUSH, &raw) == 0) {
        g_term_modified = 1;
        atexit(restore_terminal);
    }
}

/* Clear screen */
static void clear_screen(void) {
    printf("\033[2J\033[H");
    fflush(stdout);
}

/* Trim whitespace from string */
static char *trim(char *str) {
    char *end;

    while (*str == ' ' || *str == '\t' || *str == '\n' || *str == '\r')
        str++;

    if (*str == 0) return str;

    end = str + strlen(str) - 1;
    while (end > str && (*end == ' ' || *end == '\t' || *end == '\n' || *end == '\r'))
        end--;

    end[1] = '\0';
    return str;
}

/* Get the directory where the executable is located */
static int get_exe_dir(char *buf, size_t size) {
    ssize_t len = readlink("/proc/self/exe", buf, size - 1);
    if (len == -1) return -1;

    buf[len] = '\0';

    /* Remove executable name to get directory */
    char *last_slash = strrchr(buf, '/');
    if (last_slash) *last_slash = '\0';

    return 0;
}

/* Load configuration from config.ini */
static int load_config(Config *cfg) {
    char config_path[MAX_PATH_LEN];
    char line[MAX_LINE_LEN];
    char section[32] = "";
    FILE *fp;

    /* Default values */
    memset(cfg, 0, sizeof(*cfg));
    cfg->interval = 1.0f;

    /* Build config path relative to executable */
    if (get_exe_dir(config_path, sizeof(config_path)) == -1) {
        strcpy(config_path, ".");
    }
    strncat(config_path, "/" CONFIG_FILE, sizeof(config_path) - strlen(config_path) - 1);

    fp = fopen(config_path, "r");
    if (!fp) {
        /* Try current directory */
        fp = fopen(CONFIG_FILE, "r");
        if (!fp) {
            fprintf(stderr, "Error: Cannot open %s\n", CONFIG_FILE);
            return -1;
        }
    }

    while (fgets(line, sizeof(line), fp)) {
        char *trimmed = trim(line);

        /* Skip empty lines and comments */
        if (trimmed[0] == '\0' || trimmed[0] == '#' || trimmed[0] == ';')
            continue;

        /* Section header */
        if (trimmed[0] == '[') {
            char *end = strchr(trimmed, ']');
            if (end) {
                *end = '\0';
                strncpy(section, trimmed + 1, sizeof(section) - 1);
            }
            continue;
        }

        /* Key = Value */
        char *eq = strchr(trimmed, '=');
        if (!eq) continue;

        *eq = '\0';
        char *key = trim(trimmed);
        char *value = trim(eq + 1);

        if (strcmp(section, "smb") == 0) {
            if (strcmp(key, "server") == 0)
                strncpy(cfg->server, value, sizeof(cfg->server) - 1);
            else if (strcmp(key, "share") == 0)
                strncpy(cfg->share, value, sizeof(cfg->share) - 1);
            else if (strcmp(key, "path") == 0)
                strncpy(cfg->path, value, sizeof(cfg->path) - 1);
            else if (strcmp(key, "username") == 0)
                strncpy(cfg->username, value, sizeof(cfg->username) - 1);
            else if (strcmp(key, "password") == 0)
                strncpy(cfg->password, value, sizeof(cfg->password) - 1);
            else if (strcmp(key, "workgroup") == 0)
                strncpy(cfg->workgroup, value, sizeof(cfg->workgroup) - 1);
        } else if (strcmp(section, "monitor") == 0) {
            if (strcmp(key, "interval") == 0)
                cfg->interval = atof(value);
        }
    }

    fclose(fp);

    /* Validate required fields */
    if (cfg->server[0] == '\0' || cfg->share[0] == '\0' ||
        cfg->username[0] == '\0' || cfg->password[0] == '\0') {
        fprintf(stderr, "Error: Missing required SMB configuration\n");
        return -1;
    }

    return 0;
}

/* Get clipboard content using xclip */
static char *get_clipboard(void) {
    FILE *fp;
    char *content = NULL;
    size_t size = 0;
    size_t capacity = 4096;
    size_t n;

    fp = popen("xclip -selection clipboard -o 2>/dev/null", "r");
    if (!fp) return NULL;

    content = malloc(capacity);
    if (!content) {
        pclose(fp);
        return NULL;
    }

    while ((n = fread(content + size, 1, capacity - size - 1, fp)) > 0) {
        size += n;
        if (size >= capacity - 1) {
            if (capacity >= MAX_CLIPBOARD_SIZE) break;
            capacity *= 2;
            char *new_content = realloc(content, capacity);
            if (!new_content) {
                free(content);
                pclose(fp);
                return NULL;
            }
            content = new_content;
        }
    }

    content[size] = '\0';
    pclose(fp);

    /* Return NULL if empty */
    if (size == 0) {
        free(content);
        return NULL;
    }

    return content;
}

/* Get current username */
static const char *get_username(void) {
    struct passwd *pw = getpwuid(getuid());
    if (pw) return pw->pw_name;
    return "unknown";
}

/* Get current timestamp in ISO format */
static void get_timestamp(char *buf, size_t size) {
    time_t now = time(NULL);
    struct tm *tm = localtime(&now);
    strftime(buf, size, "%Y-%m-%dT%H:%M:%S", tm);
}

/* Get current time for display */
static void get_time_str(char *buf, size_t size) {
    time_t now = time(NULL);
    struct tm *tm = localtime(&now);
    strftime(buf, size, "%H:%M:%S", tm);
}

/* Create clipboard archive and upload via SMB */
static int transfer_clipboard(const char *content, const Config *cfg, char *error_msg) {
    char temp_dir[MAX_PATH_LEN];
    char data_dir[MAX_PATH_LEN];
    char meta_file[MAX_PATH_LEN];
    char content_file[MAX_PATH_LEN];
    char archive_file[MAX_PATH_LEN];
    char cmd[MAX_CMD_LEN];
    const char *username = get_username();
    char timestamp[32];
    FILE *fp;
    int ret = -1;

    error_msg[0] = '\0';

    /* Create temp directory */
    snprintf(temp_dir, sizeof(temp_dir), "/tmp/iclip_%d", getpid());
    snprintf(data_dir, sizeof(data_dir), "%s/clipboard_data", temp_dir);

    if (mkdir(temp_dir, 0700) == -1 && errno != EEXIST) {
        snprintf(error_msg, 128, "mkdir failed: %s", strerror(errno));
        return -1;
    }
    if (mkdir(data_dir, 0700) == -1 && errno != EEXIST) {
        snprintf(error_msg, 128, "mkdir failed: %s", strerror(errno));
        goto cleanup;
    }

    /* Write meta.json */
    get_timestamp(timestamp, sizeof(timestamp));
    snprintf(meta_file, sizeof(meta_file), "%s/meta.json", data_dir);
    fp = fopen(meta_file, "w");
    if (!fp) {
        snprintf(error_msg, 128, "Cannot create meta.json");
        goto cleanup;
    }
    fprintf(fp, "{\n  \"type\": \"text\",\n  \"timestamp\": \"%s\",\n"
                "  \"username\": \"%s\",\n  \"encoding\": \"utf-8\"\n}\n",
            timestamp, username);
    fclose(fp);

    /* Write content.txt */
    snprintf(content_file, sizeof(content_file), "%s/content.txt", data_dir);
    fp = fopen(content_file, "w");
    if (!fp) {
        snprintf(error_msg, 128, "Cannot create content.txt");
        goto cleanup;
    }
    fputs(content, fp);
    fclose(fp);

    /* Create tar.gz archive */
    snprintf(archive_file, sizeof(archive_file), "%s/clipboard_%s.tar.gz", temp_dir, username);
    snprintf(cmd, sizeof(cmd), "tar -czf '%s' -C '%s' clipboard_data 2>/dev/null",
             archive_file, temp_dir);
    if (system(cmd) != 0) {
        snprintf(error_msg, 128, "tar command failed");
        goto cleanup;
    }

    /* Upload via smbclient */
    if (cfg->workgroup[0]) {
        snprintf(cmd, sizeof(cmd),
                 "smbclient '//%s/%s' -U '%s%%%s' -W '%s' -c 'cd %s; put \"%s\" \"clipboard_%s.tar.gz\"' 2>&1",
                 cfg->server, cfg->share, cfg->username, cfg->password,
                 cfg->workgroup, cfg->path, archive_file, username);
    } else {
        snprintf(cmd, sizeof(cmd),
                 "smbclient '//%s/%s' -U '%s%%%s' -c 'cd %s; put \"%s\" \"clipboard_%s.tar.gz\"' 2>&1",
                 cfg->server, cfg->share, cfg->username, cfg->password,
                 cfg->path, archive_file, username);
    }

    fp = popen(cmd, "r");
    if (!fp) {
        snprintf(error_msg, 128, "smbclient failed to start");
        goto cleanup;
    }

    /* Read smbclient output for error checking */
    char output[512] = "";
    size_t total = 0;
    size_t n;
    while ((n = fread(output + total, 1, sizeof(output) - total - 1, fp)) > 0) {
        total += n;
    }
    output[total] = '\0';

    int status = pclose(fp);
    if (status != 0) {
        /* Extract first line of error */
        char *newline = strchr(output, '\n');
        if (newline) *newline = '\0';
        if (output[0])
            snprintf(error_msg, 128, "SMB: %.100s", output);
        else
            snprintf(error_msg, 128, "smbclient failed (exit %d)", WEXITSTATUS(status));
        goto cleanup;
    }

    ret = 0;

cleanup:
    /* Remove temp files */
    snprintf(cmd, sizeof(cmd), "rm -rf '%s'", temp_dir);
    (void)system(cmd);

    return ret;
}

/* Truncate string for display */
static void truncate_for_display(const char *src, char *dst, size_t max_len) {
    size_t i, j = 0;

    for (i = 0; src[i] && j < max_len - 4; i++) {
        if (src[i] == '\n') {
            dst[j++] = '\\';
            dst[j++] = 'n';
        } else if (src[i] == '\t') {
            dst[j++] = '\\';
            dst[j++] = 't';
        } else if ((unsigned char)src[i] >= 32) {
            dst[j++] = src[i];
        }
    }

    if (src[i]) {
        dst[j++] = '.';
        dst[j++] = '.';
        dst[j++] = '.';
    }
    dst[j] = '\0';
}

/* Render and display UI */
static void display_ui(const UIState *state) {
    clear_screen();

    printf("╔══════════════════════════════════════╗\n");
    printf("║         iclip_client v%-14s ║\n", VERSION);
    printf("╠══════════════════════════════════════╣\n");
    printf("║ Status: %s %-26s ║\n",
           state->paused ? "○" : "●",
           state->paused ? "Paused" : state->status);
    printf("║ Last sync: %-26s ║\n", state->last_sync[0] ? state->last_sync : "Never");
    printf("╠══════════════════════════════════════╣\n");
    printf("║ Clipboard content:                   ║\n");
    printf("║ ┌────────────────────────────────┐   ║\n");
    printf("║ │ %-32.32s │   ║\n", state->clipboard_preview[0] ? state->clipboard_preview : "(empty)");
    printf("║ └────────────────────────────────┘   ║\n");
    printf("╠══════════════════════════════════════╣\n");

    if (state->error_msg[0]) {
        printf("║ Error: %-30.30s ║\n", state->error_msg);
        printf("╠══════════════════════════════════════╣\n");
    }

    printf("║ [q] Quit  [p] Pause  [r] Refresh     ║\n");
    printf("╚══════════════════════════════════════╝\n");

    fflush(stdout);
}

/* Check for keypress (non-blocking) */
static int check_keypress(void) {
    fd_set fds;
    struct timeval tv = {0, 0};
    char c;

    FD_ZERO(&fds);
    FD_SET(STDIN_FILENO, &fds);

    if (select(STDIN_FILENO + 1, &fds, NULL, NULL, &tv) > 0) {
        if (read(STDIN_FILENO, &c, 1) == 1) {
            return c;
        }
    }
    return -1;
}

/* Main function */
int main(int argc, char *argv[]) {
    Config cfg;
    UIState state = {0};
    char *last_content = NULL;
    char *current_content = NULL;

    (void)argc;
    (void)argv;

    /* Setup signal handlers */
    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);

    /* Load configuration */
    if (load_config(&cfg) != 0) {
        fprintf(stderr, "Please create config.ini with your SMB settings.\n");
        return 1;
    }

    printf("Starting iclip_client...\n");
    printf("SMB target: //%s/%s/%s\n", cfg.server, cfg.share, cfg.path);
    printf("Press Enter to start monitoring...\n");
    getchar();

    /* Set terminal to raw mode */
    set_raw_terminal();

    /* Initialize state */
    strcpy(state.status, "Running");
    state.paused = 0;

    /* Initial display */
    display_ui(&state);

    /* Main loop */
    while (g_running) {
        /* Check for keypress */
        int key = check_keypress();
        if (key != -1) {
            if (key == 'q' || key == 'Q') {
                break;
            } else if (key == 'p' || key == 'P') {
                state.paused = !state.paused;
                display_ui(&state);
            } else if (key == 'r' || key == 'R') {
                /* Force refresh */
                free(last_content);
                last_content = NULL;
                display_ui(&state);
            }
        }

        if (state.paused) {
            usleep(100000);  /* 100ms */
            continue;
        }

        /* Get clipboard content */
        current_content = get_clipboard();

        /* Check for changes */
        if (current_content != NULL) {
            int changed = (last_content == NULL) ||
                          (strcmp(current_content, last_content) != 0);

            if (changed) {
                /* Update preview */
                truncate_for_display(current_content, state.clipboard_preview,
                                   sizeof(state.clipboard_preview));

                /* Upload */
                strcpy(state.status, "Uploading...");
                state.error_msg[0] = '\0';
                display_ui(&state);

                char error_msg[128];
                if (transfer_clipboard(current_content, &cfg, error_msg) == 0) {
                    get_time_str(state.last_sync, sizeof(state.last_sync));
                    strcpy(state.status, "Running");
                    state.error_msg[0] = '\0';

                    /* Update last content */
                    free(last_content);
                    last_content = current_content;
                    current_content = NULL;
                } else {
                    strcpy(state.status, "Error");
                    snprintf(state.error_msg, sizeof(state.error_msg), "%s", error_msg);
                }

                display_ui(&state);
            }
        }

        free(current_content);
        current_content = NULL;

        /* Sleep for interval */
        usleep((useconds_t)(cfg.interval * 1000000));
    }

    /* Cleanup */
    free(last_content);
    restore_terminal();

    printf("\nGoodbye!\n");
    return 0;
}
