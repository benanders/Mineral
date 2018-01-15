#version 330

uniform vec4 sunriseColor;

in float fragAlpha;
out vec4 color;

void main() {
	// The alpha multiplier is 1 for the centre point of the fan, and 0 for all
	// the points on the rim. This means the sunrise color fades to nothing on
	// the rim, and has an alpha component of sunriseColor.a at the centre
	color = vec4(sunriseColor.rgb, sunriseColor.a * fragAlpha);
}
