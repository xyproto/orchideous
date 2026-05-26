// Box2D 3.x - Falling boxes simulation
//
// Demonstrates the Box2D 3 C API: world creation, static ground,
// dynamic bodies, and stepping the simulation.

#include <box2d/box2d.h>
#include <cstdio>

int main()
{
    // World with standard gravity
    b2WorldDef worldDef = b2DefaultWorldDef();
    worldDef.gravity = { 0.0f, -10.0f };
    b2WorldId world = b2CreateWorld(&worldDef);

    // Static ground body: a wide box at y = -1
    b2BodyDef groundBodyDef = b2DefaultBodyDef();
    groundBodyDef.position = { 0.0f, -1.0f };
    b2BodyId ground = b2CreateBody(world, &groundBodyDef);

    b2Polygon groundBox = b2MakeBox(50.0f, 1.0f);
    b2ShapeDef groundShapeDef = b2DefaultShapeDef();
    b2CreatePolygonShape(ground, &groundShapeDef, &groundBox);

    // Dynamic body: a small box dropped from height
    b2BodyDef bodyDef = b2DefaultBodyDef();
    bodyDef.type = b2_dynamicBody;
    bodyDef.position = { 0.0f, 10.0f };
    b2BodyId body = b2CreateBody(world, &bodyDef);

    b2Polygon dynamicBox = b2MakeBox(0.5f, 0.5f);
    b2ShapeDef shapeDef = b2DefaultShapeDef();
    shapeDef.density = 1.0f;
    shapeDef.material.friction = 0.3f;
    shapeDef.material.restitution = 0.5f;
    b2CreatePolygonShape(body, &shapeDef, &dynamicBox);

    // Simulate 3 seconds at 60 Hz
    float timeStep = 1.0f / 60.0f;
    int subSteps = 4;
    int steps = 180;

    std::printf("Simulating a box falling from y=10 onto ground at y=0\n\n");
    std::printf("%5s  %8s  %8s  %8s\n", "step", "x", "y", "angle");

    for (int i = 0; i < steps; ++i) {
        b2World_Step(world, timeStep, subSteps);

        // Print position every 15 steps (~4 times per second)
        if (i % 15 == 0) {
            b2Vec2 pos = b2Body_GetPosition(body);
            b2Rot rot = b2Body_GetRotation(body);
            float angle = b2Rot_GetAngle(rot);
            std::printf("%5d  %8.4f  %8.4f  %8.4f\n", i, pos.x, pos.y, angle);
        }
    }

    b2Vec2 finalPos = b2Body_GetPosition(body);
    std::printf("\nFinal position: (%.4f, %.4f)\n", finalPos.x, finalPos.y);

    b2DestroyWorld(world);
    return 0;
}
