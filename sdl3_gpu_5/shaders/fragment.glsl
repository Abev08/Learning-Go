#version 460

layout (set = 2, binding = 0) uniform sampler2D tex_sampler;

layout (location = 0) in vec4 color;
layout (location = 1) in vec2 uv;

layout (location = 0) out vec4 out_color;

void main()
{
  out_color = texture(tex_sampler, uv) * color;
  if (out_color.a == 0) {
    discard;
  }
}