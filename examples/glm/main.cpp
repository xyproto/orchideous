// GLM - Common game-dev matrix and vector operations
//
// Demonstrates perspective projection, view matrices, and
// model transforms using GLM (OpenGL Mathematics).

#define GLM_FORCE_RADIANS
#define GLM_FORCE_DEPTH_ZERO_TO_ONE
#include <glm/glm.hpp>
#include <glm/gtc/matrix_transform.hpp>
#include <glm/gtc/type_ptr.hpp>

#include <cmath>
#include <iomanip>
#include <iostream>

static void print_mat4(const char* label, const glm::mat4& m)
{
    std::cout << label << ":\n";
    const float* p = glm::value_ptr(m);
    for (int row = 0; row < 4; ++row) {
        std::cout << "  [";
        for (int col = 0; col < 4; ++col) {
            std::cout << std::setw(9) << std::fixed << std::setprecision(4) << p[col * 4 + row];
            if (col < 3)
                std::cout << ", ";
        }
        std::cout << "]\n";
    }
}

static void print_vec4(const char* label, const glm::vec4& v)
{
    std::cout << label << ": ("
              << std::fixed << std::setprecision(4)
              << v.x << ", " << v.y << ", " << v.z << ", " << v.w << ")\n";
}

int main()
{
    // Perspective projection: 60° FOV, 16:9, near=0.1, far=100
    float aspect = 16.0f / 9.0f;
    glm::mat4 proj = glm::perspective(glm::radians(60.0f), aspect, 0.1f, 100.0f);
    print_mat4("Perspective projection (60° FOV, 16:9)", proj);

    // Camera at (0, 2, 5) looking at origin, Y-up
    glm::mat4 view = glm::lookAt(
        glm::vec3(0.0f, 2.0f, 5.0f),
        glm::vec3(0.0f, 0.0f, 0.0f),
        glm::vec3(0.0f, 1.0f, 0.0f));
    print_mat4("\nView matrix (camera at 0,2,5)", view);

    // Model: translate (1, 0, -3) then rotate 45° around Y
    glm::mat4 model = glm::mat4(1.0f);
    model = glm::translate(model, glm::vec3(1.0f, 0.0f, -3.0f));
    model = glm::rotate(model, glm::radians(45.0f), glm::vec3(0.0f, 1.0f, 0.0f));
    print_mat4("\nModel matrix (translate + rotate 45° Y)", model);

    // Full MVP
    glm::mat4 mvp = proj * view * model;
    print_mat4("\nMVP (combined)", mvp);

    // Transform a point through the pipeline
    glm::vec4 world_pos(0.0f, 1.0f, 0.0f, 1.0f);
    glm::vec4 clip_pos = mvp * world_pos;
    print_vec4("\nWorld position", world_pos);
    print_vec4("Clip position (after MVP)", clip_pos);

    // Perspective divide → NDC
    glm::vec3 ndc = glm::vec3(clip_pos) / clip_pos.w;
    std::cout << "NDC: (" << ndc.x << ", " << ndc.y << ", " << ndc.z << ")\n";

    return 0;
}
