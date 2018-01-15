#version 330

in vec2 fragUV;
out vec4 color;

void main() {
	color = vec4(fragUV, 0.0, 1.0);
}
