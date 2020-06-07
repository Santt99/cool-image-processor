#include <stdint.h>
#include <iostream>

#define STB_IMAGE_IMPLEMENTATION
#include "stb_image.h"

using namespace std;
int main()
{
    int width, height, bpp;
    int desire_channels = 3;
    uint8_t *rgb_image = stbi_load("image.jpg", &width, &height, &bpp, desire_channels);
    cout << bpp;

    uint8_t *pixel = rgb_image;
    for (int i = 0; i < height; ++i)
    {
        for (int j = 0; j < width; ++j, pixel += desire_channels)
        {
            cout << (int)(pixel[0]);
            // Do something with r, g, b
        }
    }

    //Rewrite the image to make sure that image reader is working correctly

    stbi_image_free(rgb_image);

    return 0;
}