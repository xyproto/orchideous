#include <cstdio>
#include <filesystem>
#include <iostream>
#include <string>
#include <vte/vte.h>

/*
 * A terminal emulator only for playing Dunnet.
 * Inspired by: https://vincent.bernat.ch/en/blog/2017-write-own-terminal
 */

using namespace std::string_literals;

static VteTerminal* terminal;
static GtkWindow* window;

void spawn_callback(VteTerminal*, GPid, GError* error, gpointer)
{
    if (error) {
        std::cerr << "spawn error: " << error->message << std::endl;
        g_error_free(error);
    }
}

void on_child_exited(VteTerminal*, int, gpointer data)
{
    auto app = G_APPLICATION(data);
    g_application_quit(app);
}

void on_close_request(GtkWindow*, gpointer data)
{
    auto app = G_APPLICATION(data);
    g_application_quit(app);
}

static void activate(GtkApplication* app, gpointer)
{
    // Create the main window
    window = GTK_WINDOW(gtk_application_window_new(app));
    gtk_window_set_title(window, "Dunnet");

    // Create the terminal
    auto term_widget = vte_terminal_new();
    terminal = VTE_TERMINAL(term_widget);

    using std::filesystem::exists;
    using std::filesystem::path;
    using std::filesystem::perms;
    using std::filesystem::status;

    // Search for the emacs executable in $PATH
    char* dup = strdup(getenv("PATH"));
    char* s = dup;
    char* p = nullptr;
    path found { "emacs"s };
    do {
        p = strchr(s, ':');
        if (p != nullptr) {
            p[0] = 0;
        }
        if (exists(path(s) / found)) {
            found = path(s) / found;
            break;
        }
        s = p + 1;
    } while (p != nullptr);
    free(dup);

    if (found == "emacs"s) {
        std::cerr << found << " does not exist in PATH" << std::endl;
        g_application_quit(G_APPLICATION(app));
        return;
    }

    // Build the command
    const char* command[5];
    command[0] = found.c_str();
    command[1] = "-batch";
    command[2] = "-l";
    command[3] = "dunnet";
    command[4] = nullptr;

    // Check if the executable is executable
    const auto perm = status(command[0]).permissions();
    if ((perm & perms::owner_exec) == perms::none) {
        std::cerr << command[0] << " is not executable for this user" << std::endl;
        g_application_quit(G_APPLICATION(app));
        return;
    }

    // Spawn the terminal asynchronously (GTK4 VTE has no spawn_sync)
    vte_terminal_spawn_async(terminal, VTE_PTY_DEFAULT,
        nullptr, // working directory
        (char**)command, // command
        nullptr, // environment
        (GSpawnFlags)0, // spawn flags
        nullptr, nullptr, nullptr, // child setup
        -1, // timeout
        nullptr, // cancellable
        spawn_callback, // callback
        nullptr); // user data

    // Set background color to 95% opaque black
    GdkRGBA black;
    gdk_rgba_parse(&black, "rgba(0, 0, 0, 0.95)");
    vte_terminal_set_color_background(terminal, &black);

    // Set foreground color
    GdkRGBA green;
    gdk_rgba_parse(&green, "chartreuse");
    vte_terminal_set_color_foreground(terminal, &green);

    // Set font
    auto font_desc = pango_font_description_from_string("courier bold 16");
    vte_terminal_set_font(terminal, font_desc);
    pango_font_description_free(font_desc);

    // Set cursor shape to UNDERLINE
    vte_terminal_set_cursor_shape(terminal, VTE_CURSOR_SHAPE_UNDERLINE);

    // Set cursor blink to OFF
    vte_terminal_set_cursor_blink_mode(terminal, VTE_CURSOR_BLINK_OFF);

    // Connect signals
    g_signal_connect(window, "close-request", G_CALLBACK(on_close_request), app);
    g_signal_connect(terminal, "child-exited", G_CALLBACK(on_child_exited), app);

    // Add the terminal to the window
    gtk_window_set_child(window, term_widget);

    // Present the window (GTK4: widgets are visible by default)
    gtk_window_present(window);
}

int main(int argc, char* argv[])
{
    auto app = gtk_application_new("com.example.dunnet", G_APPLICATION_DEFAULT_FLAGS);
    g_signal_connect(app, "activate", G_CALLBACK(activate), nullptr);
    int status = g_application_run(G_APPLICATION(app), argc, argv);
    g_object_unref(app);
    return status;
}
