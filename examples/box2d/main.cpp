// Box2D - Falling box simulation
//
// Demonstrates Box2D physics: world creation, static ground, dynamic body,
// and stepping the simulation. Supports both the 2.4 C++ API and the 3.x C API.

#include <box2d/box2d.h>
#include <cstdio>

#if __has_include(<box2d/b2_world.h>)
#define BOX2D_V2
#endif

int main()
{
    std::printf("Simulating a box falling from y=10 onto ground at y=0\n\n");
    std::printf("%5s  %8s  %8s  %8s\n", "step", "x", "y", "angle");

    float timeStep = 1.0f / 60.0f;
    int steps = 180;

#ifdef BOX2D_V2
    // --- Box2D 2.4 C++ API ---

    b2Vec2 gravity(0.0f, -10.0f);
    b2World world(gravity);

    // Static ground body
    b2BodyDef groundBodyDef;
    groundBodyDef.position.Set(0.0f, -1.0f);
    b2Body* ground = world.CreateBody(&groundBodyDef);

    b2PolygonShape groundBox;
    groundBox.SetAsBox(50.0f, 1.0f);
    ground->CreateFixture(&groundBox, 0.0f);

    // Dynamic body dropped from height
    b2BodyDef bodyDef;
    bodyDef.type = b2_dynamicBody;
    bodyDef.position.Set(0.0f, 10.0f);
    b2Body* body = world.CreateBody(&bodyDef);

    b2PolygonShape dynamicBox;
    dynamicBox.SetAsBox(0.5f, 0.5f);

    b2FixtureDef fixtureDef;
    fixtureDef.shape = &dynamicBox;
    fixtureDef.density = 1.0f;
    fixtureDef.friction = 0.3f;
    fixtureDef.restitution = 0.5f;
    body->CreateFixture(&fixtureDef);

    int velocityIters = 6;
    int positionIters = 2;

    for (int i = 0; i < steps; ++i) {
        world.Step(timeStep, velocityIters, positionIters);
        if (i % 15 == 0) {
            b2Vec2 pos = body->GetPosition();
            float angle = body->GetAngle();
            std::printf("%5d  %8.4f  %8.4f  %8.4f\n", i, pos.x, pos.y, angle);
        }
    }

    b2Vec2 finalPos = body->GetPosition();
    std::printf("\nFinal position: (%.4f, %.4f)\n", finalPos.x, finalPos.y);

#else
    // --- Box2D 3.x C API ---

    b2WorldDef worldDef = b2DefaultWorldDef();
    worldDef.gravity = {0.0f, -10.0f};
    b2WorldId world = b2CreateWorld(&worldDef);

    // Static ground body
    b2BodyDef groundBodyDef = b2DefaultBodyDef();
    groundBodyDef.position = {0.0f, -1.0f};
    b2BodyId ground = b2CreateBody(world, &groundBodyDef);

    b2Polygon groundBox = b2MakeBox(50.0f, 1.0f);
    b2ShapeDef groundShapeDef = b2DefaultShapeDef();
    b2CreatePolygonShape(ground, &groundShapeDef, &groundBox);

    // Dynamic body dropped from height
    b2BodyDef bodyDef = b2DefaultBodyDef();
    bodyDef.type = b2_dynamicBody;
    bodyDef.position = {0.0f, 10.0f};
    b2BodyId body = b2CreateBody(world, &bodyDef);

    b2Polygon dynamicBox = b2MakeBox(0.5f, 0.5f);
    b2ShapeDef shapeDef = b2DefaultShapeDef();
    shapeDef.density = 1.0f;
    shapeDef.material.friction = 0.3f;
    shapeDef.material.restitution = 0.5f;
    b2CreatePolygonShape(body, &shapeDef, &dynamicBox);

    int subSteps = 4;

    for (int i = 0; i < steps; ++i) {
        b2World_Step(world, timeStep, subSteps);
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
#endif

    return 0;
}
