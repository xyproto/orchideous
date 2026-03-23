// Based on the "Basic OpenGL ES 3 + SDL3 template code" example from the book
// "Getting Started with OpenGL ES 3+ Programming: Learn Modern OpenGL Basics" by Hans de Ruiter

#include <GLES3/gl3.h>
#include <SDL3/SDL.h>
#include <SDL3/SDL_opengles2.h>

#include <cstdlib>

const unsigned int DISP_WIDTH = 640;
const unsigned int DISP_HEIGHT = 480;

int main(int argc, char* args[])
{
    // The window
    SDL_Window* window = nullptr;

    // The OpenGL context
    SDL_GLContext context = nullptr;

    // Init SDL
    if (!SDL_Init(SDL_INIT_VIDEO)) {
        SDL_Log("SDL could not initialize! SDL_Error: %s\n", SDL_GetError());
        return EXIT_FAILURE;
    }

    // Request OpenGL ES 3.0
    SDL_GL_SetAttribute(SDL_GL_CONTEXT_PROFILE_MASK, SDL_GL_CONTEXT_PROFILE_ES);
    SDL_GL_SetAttribute(SDL_GL_CONTEXT_MAJOR_VERSION, 3);
    SDL_GL_SetAttribute(SDL_GL_CONTEXT_MINOR_VERSION, 0);

    // Enable double-buffering
    SDL_GL_SetAttribute(SDL_GL_DOUBLEBUFFER, 1);

    // Create the window
    window = SDL_CreateWindow("SDL 3 + OpenGL ES 3.0", DISP_WIDTH, DISP_HEIGHT, SDL_WINDOW_OPENGL);
    if (!window) {
        SDL_ShowSimpleMessageBox(
            SDL_MESSAGEBOX_ERROR, "Error", "Couldn't create the main window.", nullptr);
        return EXIT_FAILURE;
    }

    context = SDL_GL_CreateContext(window);
    if (!context) {
        SDL_ShowSimpleMessageBox(
            SDL_MESSAGEBOX_ERROR, "Error", "Couldn't create an OpenGL context.", nullptr);
        return EXIT_FAILURE;
    }

    // Clear to a blue color
    glClearColor(0.4f, 0.7f, 1.0f, 1.0f);
    glClear(GL_COLOR_BUFFER_BIT);

    // Update the window
    SDL_GL_SwapWindow(window);

    // Wait for the user to quit
    bool quit = false;
    while (!quit) {
        SDL_Event event;
        if (SDL_WaitEvent(&event)) {
            if (event.type == SDL_EVENT_QUIT) {
                // User wants to quit
                quit = true;
            }
        }
    }

    SDL_GL_DestroyContext(context);
    SDL_DestroyWindow(window);
    SDL_Quit();
    return EXIT_SUCCESS;
}
