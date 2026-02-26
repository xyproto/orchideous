#include <SFML/Graphics.hpp>
#include <SFML/OpenGL.hpp>
#include <array>
#include <cmath>
#include <map>

/* My second OpenGL exercise
 * Copyright (c) 1992,2019 Joel Yliluoma - https://iki.fi/bisqwit/
 * First published at: https://youtube.com/Bisqwit
 * Source code license: MIT
 * Compile with (example):
 *   g++ 75000.cc -Wall -Wextra -Ofast -std=c++20 $(pkg-config sfml-graphics --libs --cflags) -lGL
 */

static const char recipe[]
    = "lidjehfhfhhideiefedefedefekedeiefedefedefejfdeiefedefedefejeeieefed"
      "efedefeiekedefedefedefeiekedefedefedefeiefefedefedefedefeieghfhfhfhm";

int main()
{
    using namespace sf;
    // Create the main window
    RenderWindow window(VideoMode({ 3840, 2160 }), "Hello", Style::Default, State::Windowed,
        ContextSettings { .depthBits = 24, .antiAliasingLevel = 2 });
    window.setVerticalSyncEnabled(true);

    // Configure OpenGL features.
    window.resetGLStates();
    // SFML 3 no longer enables legacy fixed-function client states, so enable them manually.
    glEnableClientState(GL_VERTEX_ARRAY);
    glEnableClientState(GL_COLOR_ARRAY);
    glEnableClientState(GL_TEXTURE_COORD_ARRAY);
    glEnable(GL_TEXTURE_2D);
    glDisableClientState(GL_NORMAL_ARRAY); // Disable normals, not used.
    glEnable(GL_DEPTH_TEST); // SFML disables z-buffer. Re-enable, because we need it.
    glClearDepth(1.f);

    // Load some textures
    Texture tx[7];
    (void)tx[0].loadFromFile("resources/bottom.jpg");
    (void)tx[1].loadFromFile("resources/top.jpg");
    (void)tx[2].loadFromFile("resources/left.jpg");
    (void)tx[3].loadFromFile("resources/right.jpg");
    (void)tx[4].loadFromFile("resources/back.jpg");
    (void)tx[5].loadFromFile("resources/front.jpg");
    (void)tx[6].loadFromFile("resources/wall3.jpg");
    (void)tx[6].generateMipmap();

    // Construct the world geometry from axis-aligned cuboids made of triangles.
    std::vector<GLfloat> tri;
    auto addcuboid = [&](unsigned mask, std::array<float, 2> x, std::array<float, 2> z,
                         std::array<float, 2> y, std::array<float, 3> c, std::array<float, 3> u,
                         std::array<float, 3> v) {
        auto ext = [](auto m, unsigned n, unsigned b = 1) {
            return (m >> (n * b)) & ~(~0u << b);
        }; // extracts bits
        // Generates: For six vertices, color(rgb), coordinate(xyz) and texture coord(uv).
        std::array p { &c[0], &c[0], &c[0], &x[0], &y[0], &z[0], &u[0], &v[0] };
        // capflag(1 bit), mask(3 bits), X(4 bits), Y(4 bits), Z(4 bits), U(4 bits), V(4 bits)
        for (unsigned m : std::array { 0x960339, 0xA9F339, 0x436039, 0x4C6F39, 0x406C39,
                 0x4F6339 }) // bottom, top, four sides
            if (std::uint64_t s = (m >> 23) * 0b11'000'111 * (~0llu / 255); mask & m)
                for (unsigned n = 0; n < 6 * 8; ++n)
                    tri.push_back(
                        p[n % 8][ext(m, ext(012345444u, n % 8, 3) * 4 - ext(0123341u, n / 8, 3))
                            << ext(s, n)]);
        // 123341 = order of vertices in two triangles; 12345444 = nibble indexes in "m" for each
        // of 8 values
    };
    // Part 1: Skybox. Perfect cube. Size is mostly irrelevant, as long as it's farther than the
    // near clipping plane.
    //         "Mostly irrelevant", because its distance from viewer still influences how much fog
    //         affects it. As an easy exercise, test and see what happens if you stretch the cube!
    //         The first three {}s are the X, Y and Z extremes of the cube.
    //         It could be arbitrary rotated around Y axis, but its easiest to specify axis-aligned
    //         coordinates. The next one controls the darkness/lightness (three values are provided
    //         for top, bottom and cap respectively) And the last two control the texture offsets:
    //         min,max,and max for horizontal surfaces.
    addcuboid(
        7 << 20, { -10, 10 }, { -10, 10 }, { -10, 10 }, { 1, 1, 1 }, { 0, 1, 1 }, { 0, 1, 1 });
    // Part 2: Floor plane
    addcuboid(
        1 << 20, { -30, 30 }, { -30, 30 }, { 0, 10 }, { .3, .3, .4 }, { 0, 0, 60 }, { 0, 0, 60 });
    // Part 3: Random "buildings"
    for (int rem = 0, p = 0, z = -14; z < 15; ++z)
        for (int x = -21; x < 21; ++x) {
            if (!rem--) {
                rem = recipe[p++] - 'd';
                if (rem & 8)
                    rem += 414;
            } // RLE compression, odd/even coding
            if (float w = .5f,
                h = (p & 1) ? (std::rand() % 2) * .05f : .8f * (4 + std::rand() % 8);
                h) // Random height
                addcuboid(6 << 20, { x - w, x + w }, { z - w, z + w }, { 0, h },
                    { .2f + (rand() % 1000) * .4e-3f, 1, .4f + (h > .1f) }, { 0, 1, 1 },
                    { 0, h, 1 });
        }
    glColorPointer(3, GL_FLOAT, 8 * sizeof(GLfloat), &tri[0]);
    glVertexPointer(3, GL_FLOAT, 8 * sizeof(GLfloat), &tri[3]);
    glTexCoordPointer(2, GL_FLOAT, 8 * sizeof(GLfloat), &tri[6]);

    GLfloat near = .03f, far = 50.f;

    // Start game loop
    float rx = 0, ry = 0, rz = 0, mx = 0, my = 0, mz = 0, lx = 0, ly = -20, lz = .5, aa = .7071f,
          ab = .7071f, ac = 0, ad = 0, fog = 1;
    // Initialize transformation matrix from quaternion (not identity, so the initial view
    // direction is correct).
    GLfloat tform[16] { 1 - 2 * (ac * ac + ad * ad), 2 * (ab * ac + aa * ad),
        2 * (ab * ad - aa * ac), 0, 2 * (ab * ac - aa * ad), 1 - 2 * (ab * ab + ad * ad),
        2 * (ac * ad + aa * ab), 0, 2 * (ab * ad + aa * ac), 2 * (ac * ad - aa * ab),
        1 - 2 * (ab * ab + ac * ac), 0, 0, 0, 0, 1 };
    for (std::map<Keyboard::Key, bool> keys; window.isOpen() && !keys[Keyboard::Key::Escape];
        window.display()) {
        // Setup up the view port, the clipping planes, the aspect ratio and the field of vision
        // (FoV)
        glMatrixMode(GL_PROJECTION);
        glLoadIdentity();
        glViewport(0, 0, window.getSize().x, window.getSize().y);
        GLfloat ratio = near * window.getSize().x / window.getSize().y;
        glFrustum(-ratio, ratio, -near, near, near, far);

        // Process events
        while (const auto event = window.pollEvent()) {
            if (event->is<Event::Closed>())
                keys[Keyboard::Key::Escape] = true;
            else if (const auto* kp = event->getIf<Event::KeyPressed>())
                keys[kp->code] = true;
            else if (const auto* kr = event->getIf<Event::KeyReleased>())
                keys[kr->code] = false;
        }
        if (keys[Keyboard::Key::V]) {
            for (std::size_t p = 6 * 6 * 8; p < tri.size(); p += 8)
                if (tri[p + 4] > 0.1)
                    tri[p + 4] *= 0.95;
            fog *= 0.95;
        }

        // The input scheme is the same as in Descent, the game by Parallax Interactive.
        // Mouse input is not handled for now.
        bool up = keys[Keyboard::Key::Up] || keys[Keyboard::Key::Numpad8];
        bool down = keys[Keyboard::Key::Down] || keys[Keyboard::Key::Numpad2],
             alt = keys[Keyboard::Key::LAlt] || keys[Keyboard::Key::RAlt];
        bool left = keys[Keyboard::Key::Left] || keys[Keyboard::Key::Numpad4],
             rleft = keys[Keyboard::Key::Q] || keys[Keyboard::Key::Numpad7];
        bool right = keys[Keyboard::Key::Right] || keys[Keyboard::Key::Numpad6],
             rright = keys[Keyboard::Key::E] || keys[Keyboard::Key::Numpad9];
        bool fwd = keys[Keyboard::Key::A], sup = keys[Keyboard::Key::Subtract],
             sleft = keys[Keyboard::Key::Numpad1];
        bool back = keys[Keyboard::Key::Z], sdown = keys[Keyboard::Key::Add],
             sright = keys[Keyboard::Key::Numpad3];

        // Apply rotation delta with hysteresis: newvalue = input*eagerness +
        // oldvalue*(1-eagerness)
        rx = rx * .8f + .2f * (up - down) * !alt;
        ry = ry * .8f + .2f * (right - left) * !alt;
        rz = rz * .8f + .2f * (rright - rleft);
        if (float rlen = std::sqrt(rx * rx + ry * ry + rz * rz); rlen > 1e-3f) // Still rotating?
        {
            // Create rotation quaternion (q), relative to the current angle that the player is
            // looking towards.
            float theta = rlen * .03f, c = std::cos(theta * .5f), s = std::sin(theta * .5f) / rlen;
            auto [qa, qb, qc, qd]
                = std::array { c, s * (tform[0] * rx + tform[1] * ry + tform[2] * rz),
                      s * (tform[4] * rx + tform[5] * ry + tform[6] * rz),
                      s * (tform[8] * rx + tform[9] * ry + tform[10] * rz) };
            // Update player angle (a) by multiplying it by the rotation quaternion (r)
            std::tie(aa, ab, ac, ad) = std::tuple { qa * aa - qb * ab - qc * ac - qd * ad,
                qb * aa + qa * ab + qd * ac - qc * ad, qc * aa - qd * ab + qa * ac + qb * ad,
                qd * aa + qc * ab - qb * ac + qa * ad };
            // Normalize to prevent floating point inaccuracies creeping in, eventually wonkifying
            // the rotations.
            std::tie(aa, ab, ac, ad)
                = std::tuple { aa * (1.f / std::sqrt(aa * aa + ab * ab + ac * ac + ad * ad)),
                      ab * (1.f / std::sqrt(aa * aa + ab * ab + ac * ac + ad * ad)),
                      ac * (1.f / std::sqrt(aa * aa + ab * ab + ac * ac + ad * ad)),
                      ad * (1.f / std::sqrt(aa * aa + ab * ab + ac * ac + ad * ad)) };
            // Recalculate the rotation matrix from the new player angle (a).
            tform[0] = 1 - 2 * (ac * ac + ad * ad);
            tform[1] = 2 * (ab * ac + aa * ad);
            tform[2] = 2 * (ab * ad - aa * ac);
            tform[4] = 2 * (ab * ac - aa * ad);
            tform[5] = 1 - 2 * (ab * ab + ad * ad);
            tform[6] = 2 * (ac * ad + aa * ab);
            tform[8] = 2 * (ab * ad + aa * ac);
            tform[9] = 2 * (ac * ad - aa * ab);
            tform[10] = 1 - 2 * (ab * ab + ac * ac);
            // Note: The above cos() and sin() were the ONLY trigonometric calculations
            //       in this rotation code. And they only control the rate of turning,
            //       not the angle of turning. You could replace them with compile-time
            //       constants and the only downside would be a constant rate of turning
            //       (either on or off). Such is the power of quaternions.
        }

        // Apply player movement delta with hysteresis
        float Mx = (sleft || (alt && left)) - (sright || (alt && right));
        float My = (sdown || (alt && down)) - (sup || (alt && up));
        float Mz = fwd - back;
        float mlen = std::sqrt(Mx * Mx + My * My + Mz * Mz) / 0.07;
        if (mlen < 1e-3f)
            mlen = 1;
        // The new movement is relative to the angle that player is looking towards.
        mx = mx * .9f + .1f * (tform[0] * Mx + tform[1] * My + tform[2] * Mz) / mlen;
        my = my * .9f + .1f * (tform[4] * Mx + tform[5] * My + tform[6] * Mz) / mlen;
        mz = mz * .9f + .1f * (tform[8] * Mx + tform[9] * My + tform[10] * Mz) / mlen;
        // Update player position (l) by the movement vector (m)
        lx += mx;
        ly += my;
        lz += mz;
        // Note: We don't do any clipping here. Player is free to move through walls and floors.
        //       Adding collision checks is a quite a complex topic, a good example of the 80/20
        //       rule, especially if you want to do it properly and have the character slide off
        //       the surface etc. So, I neglected that in favor of brevity.

        // Set up fog.
        glEnable(GL_FOG);
        glFogi(GL_FOG_MODE, GL_EXP);
        glFogfv(GL_FOG_COLOR, &std::array { .5f, .51f, .54f }[0]);
        glFogf(GL_FOG_DENSITY, fog / far);

        // Instruct OpenGL about the view rotation.
        glMatrixMode(GL_MODELVIEW);
        glLoadMatrixf(tform);

        // Render the skybox without zbuffer. View is still in origo, where the skybox is centered.
        glClear(GL_DEPTH_BUFFER_BIT);
        glDepthMask(GL_FALSE);
        for (unsigned n = 0; n < 6; ++n) {
            Texture::bind(&tx[n]);
            glDrawArrays(GL_TRIANGLES, n * 6, 6);
        }

        // After the skybox has been rendered, add the player coordinate & turn zbuffer on.
        glTranslatef(lx, ly, lz);
        glDepthMask(GL_TRUE);

        // Render everything else using a single repeated texture.
        Texture::bind(&tx[6]);
        glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_REPEAT);
        glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_REPEAT);
        glDrawArrays(GL_TRIANGLES, 6 * 6, tri.size() / 8 - 6 * 6);
    }
}
