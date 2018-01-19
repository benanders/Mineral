#version 330

uniform mat4 mvp;
uniform vec4 sunriseColor;

in vec3 position;
in float alpha;
out float fragAlpha;

void main() {
	// Modulate the z component of the position by the alpha component of the
	// sunrise color
	gl_Position = mvp * vec4(position.xy, position.z * sunriseColor.a, 1.0);
	fragAlpha = alpha;
}
