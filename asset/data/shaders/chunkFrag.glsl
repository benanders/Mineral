#version 330

uniform sampler2D blockAtlas;

in vec2 fragUV;
out vec4 color;

void main() {
	color = texture(blockAtlas, fragUV);
}
