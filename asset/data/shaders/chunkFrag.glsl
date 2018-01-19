#version 330

uniform sampler2D terrain;

in vec2 fragUV;
out vec4 color;

void main() {
	color = texture(terrain, fragUV);
}
