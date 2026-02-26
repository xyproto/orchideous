/*
 * Raylib 5 — Bouncing circles with collision explosions
 *
 * A modern C++ replacement for the old EntityX + SFML2 example.
 * Uses a data-oriented struct-of-arrays layout instead of a
 * heavyweight ECS framework.
 *
 * 500 coloured circles bounce around the screen. When two collide
 * they are destroyed and replaced by a shower of fading particles.
 * An FPS / entity counter is drawn in the top-left corner.
 */

#include <algorithm>
#include <cmath>
#include <cstdlib>
#include <ctime>
#include <raylib.h>
#include <string>
#include <vector>

// Random float in [lo, hi)
static float randf(float lo, float hi)
{
    return lo + static_cast<float>(std::rand()) / static_cast<float>(RAND_MAX) * (hi - lo);
}

// ── Components stored as parallel vectors (struct-of-arrays) ────────────────

struct Circle {
    Vector2 pos;
    Vector2 vel;
    float radius;
    Color color;
    float alpha; // fade-in: 0 → 1
};

struct Particle {
    Vector2 pos;
    Vector2 vel;
    float rotation;
    float rotationd;
    float radius;
    Color color;
    float alpha;
    float decay; // alpha units lost per second
};

// ── World state ─────────────────────────────────────────────────────────────

static std::vector<Circle> circles;
static std::vector<Particle> particles;
static constexpr int TARGET_CIRCLES = 500;

// ── Systems ─────────────────────────────────────────────────────────────────

static void spawn_circles(int screenW, int screenH)
{
    while (static_cast<int>(circles.size()) < TARGET_CIRCLES) {
        float r = randf(5, 15);
        circles.push_back(Circle { { randf(r, screenW - r), randf(r, screenH - r) },
            { randf(-100, 100), randf(-100, 100) }, r,
            { static_cast<unsigned char>(std::rand() % 128 + 127),
                static_cast<unsigned char>(std::rand() % 128 + 127),
                static_cast<unsigned char>(std::rand() % 128 + 127), 0 },
            0.0f });
    }
}

static void move_circles(float dt)
{
    for (auto& c : circles) {
        c.pos.x += c.vel.x * dt;
        c.pos.y += c.vel.y * dt;
        c.alpha = std::min(1.0f, c.alpha + dt);
    }
}

static void bounce_circles(int screenW, int screenH)
{
    for (auto& c : circles) {
        if (c.pos.x - c.radius < 0 || c.pos.x + c.radius > screenW)
            c.vel.x = -c.vel.x;
        if (c.pos.y - c.radius < 0 || c.pos.y + c.radius > screenH)
            c.vel.y = -c.vel.y;
        c.pos.x = std::clamp(c.pos.x, c.radius, static_cast<float>(screenW) - c.radius);
        c.pos.y = std::clamp(c.pos.y, c.radius, static_cast<float>(screenH) - c.radius);
    }
}

static void emit_particles(const Circle& c)
{
    float area = (PI * c.radius * c.radius) / 3.0f;
    int count = static_cast<int>(area);
    for (int i = 0; i < count; ++i) {
        float angle = randf(0, 2 * PI);
        float offset = randf(1, c.radius);
        float r = randf(1, 4);
        float rotd = randf(180, 720);
        if (std::rand() % 2)
            rotd = -rotd;

        Color col = c.color;
        col.a = 200;
        particles.push_back(
            Particle { { c.pos.x + offset * std::cos(angle), c.pos.y + offset * std::sin(angle) },
                { c.vel.x + offset * 2 * std::cos(angle), c.vel.y + offset * 2 * std::sin(angle) },
                0.0f, rotd, r, col, 200.0f, 200.0f / (r / 2.0f) });
    }
}

static void detect_collisions()
{
    std::vector<bool> dead(circles.size(), false);
    for (std::size_t i = 0; i < circles.size(); ++i) {
        if (dead[i])
            continue;
        for (std::size_t j = i + 1; j < circles.size(); ++j) {
            if (dead[j])
                continue;
            float dx = circles[i].pos.x - circles[j].pos.x;
            float dy = circles[i].pos.y - circles[j].pos.y;
            float dist = std::sqrt(dx * dx + dy * dy);
            if (dist < circles[i].radius + circles[j].radius) {
                emit_particles(circles[i]);
                emit_particles(circles[j]);
                dead[i] = dead[j] = true;
            }
        }
    }
    // Remove dead circles (stable, back-to-front)
    for (auto it = static_cast<int>(circles.size()) - 1; it >= 0; --it)
        if (dead[it])
            circles.erase(circles.begin() + it);
}

static void update_particles(float dt)
{
    for (auto& p : particles) {
        p.pos.x += p.vel.x * dt;
        p.pos.y += p.vel.y * dt;
        p.rotation += p.rotationd * dt;
        p.alpha -= p.decay * dt;
    }
    std::erase_if(particles, [](const Particle& p) { return p.alpha <= 0; });
}

static void draw_particles()
{
    for (const auto& p : particles) {
        Color col = p.color;
        col.a = static_cast<unsigned char>(std::clamp(p.alpha, 0.0f, 255.0f));
        // Draw a small rotated rectangle as a particle
        Rectangle rec { p.pos.x, p.pos.y, p.radius * 2, p.radius * 2 };
        DrawRectanglePro(rec, { p.radius, p.radius }, p.rotation, col);
    }
}

static void draw_circles()
{
    for (const auto& c : circles) {
        Color col = c.color;
        col.a = static_cast<unsigned char>(c.alpha * 255);
        DrawCircleV(c.pos, c.radius, col);
    }
}

// ── Main ────────────────────────────────────────────────────────────────────

int main()
{
    std::srand(static_cast<unsigned>(std::time(nullptr)));

    SetConfigFlags(FLAG_WINDOW_RESIZABLE | FLAG_VSYNC_HINT);
    InitWindow(1280, 720, "Raylib 5 — Bouncing Circles");

    while (!WindowShouldClose()) {
        int sw = GetScreenWidth();
        int sh = GetScreenHeight();
        float dt = GetFrameTime();

        spawn_circles(sw, sh);
        move_circles(dt);
        bounce_circles(sw, sh);
        detect_collisions();
        update_particles(dt);

        BeginDrawing();
        ClearBackground(BLACK);
        draw_particles();
        draw_circles();

        std::string info = std::to_string(circles.size() + particles.size()) + " entities ("
            + std::to_string(GetFPS()) + " fps)";
        DrawText(info.c_str(), 4, 4, 20, WHITE);
        EndDrawing();
    }

    CloseWindow();
}
