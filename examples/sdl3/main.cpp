// SDL3 - Bouncing ball with the 2D rendccerer
//
// Demonstrates the SDL3 API: window creation, event handling,
// and hardware-accelerated 2D rendering without OpenGL.

#include <SDL3/SDL.h>
#include <SDL3/SDL_main.h>

#include <algorithm>
#include <cmath>
#include <cstdlib>

static constexpr int WIDTH = 800;
static constexpr int HEIGHT = 600;

int main(int argc, char* argv[])
{
    (void)argc;
    (void)argv;

    if (!SDL_Init(SDL_INIT_VIDEO)) {
        SDL_Log("SDL_Init failed: %s", SDL_GetError());
        return 1;
    }

    SDL_Window* window = SDL_CreateWindow("SDL3 - Bouncing Ball", WIDTH, HEIGHT, SDL_WINDOW_RESIZABLE);
    if (!window) {
        SDL_Log("SDL_CreateWindow failed: %s", SDL_GetError());
        return 1;
    }

    SDL_Renderer* renderer = SDL_CreateRenderer(window, nullptr);
    if (!renderer) {
        SDL_Log("SDL_CreateRenderer failed: %s", SDL_GetError());
        return 1;
    }

    // Ball state
    float x = WIDTH / 2.0f;
    float y = HEIGHT / 2.0f;
    float vx = 320.0f;
    float vy = 240.0f;
    float radius = 20.0f;

    bool running = true;
    Uint64 last = SDL_GetTicksNS();

    while (running) {
        SDL_Event ev;
        while (SDL_PollEvent(&ev)) {
            if (ev.type == SDL_EVENT_QUIT)
                running = false;
            if (ev.type == SDL_EVENT_KEY_DOWN && ev.key.key == SDLK_ESCAPE)
                running = false;
        }

        // Delta time in seconds
        Uint64 now = SDL_GetTicksNS();
        float dt = static_cast<float>(now - last) / 1.0e9f;
        last = now;

        // Get current window size (handles resize)
        int w, h;
        SDL_GetWindowSize(window, &w, &h);

        // Move and bounce
        x += vx * dt;
        y += vy * dt;
        if (x - radius < 0 || x + radius > w) {
            vx = -vx;
            x = std::clamp(x, radius, static_cast<float>(w) - radius);
        }
        if (y - radius < 0 || y + radius > h) {
            vy = -vy;
            y = std::clamp(y, radius, static_cast<float>(h) - radius);
        }

        // Draw
        SDL_SetRenderDrawColor(renderer, 20, 20, 30, 255);
        SDL_RenderClear(renderer);

        // Draw filled circle as a series of horizontal lines
        SDL_SetRenderDrawColor(renderer, 80, 200, 255, 255);
        for (int dy = -static_cast<int>(radius); dy <= static_cast<int>(radius); ++dy) {
            int dx = static_cast<int>(std::sqrt(radius * radius - dy * dy));
            SDL_RenderLine(renderer, x - dx, y + dy, x + dx, y + dy);
        }

        SDL_RenderPresent(renderer);
    }

    SDL_DestroyRenderer(renderer);
    SDL_DestroyWindow(window);
    SDL_Quit();
    return 0;
}
