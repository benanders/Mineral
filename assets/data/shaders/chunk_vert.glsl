#version 330

uniform mat4 mvp;

in vec3 position;
in vec3 normal;
in vec2 uv;

out vec2 fragUV;

void main() {
	gl_Position = mvp * vec4(position, 1.0);
	fragUV = uv;
}
