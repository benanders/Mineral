#version 330

uniform vec3 skyColor;
uniform vec3 fogColor;
uniform float farPlane;

in vec3 fragPos;
out vec4 color;

void main() {
	// Use the position of the fragment to calculate the fog strength
	float fog_strength = length(fragPos) / (farPlane * 0.8);
	fog_strength = clamp(fog_strength, 0.0, 1.0);

	// Modulate between the sky and fog colors by the fog strength factor
	color = vec4(mix(skyColor, fogColor, fog_strength), 1.0);
}
