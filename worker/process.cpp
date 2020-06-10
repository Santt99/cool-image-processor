#include <stdint.h>
#include <iostream>
#include "images.h"

#define STB_IMAGE_IMPLEMENTATION
#include "stb_image.h"

#define STB_IMAGE_WRITE_IMPLEMENTATION
#include "stb_image_write.h"

using namespace std;
int main()
{
    int width, height, bpp;
    int desire_channels = 3;
    uint8_t *rgb_image = stbi_load("image.jpg", &width, &height, &bpp, desire_channels);
    cout << bpp;

    unsigned char *data = (unsigned char *) malloc(sizeof(unsigned char) * width*height*3);
    unsigned char *out = (unsigned char *) malloc(sizeof(unsigned char) * width*height*3);
    uint8_t *pixel = rgb_image;
    int index = 0; 
    
    for (int i = 0; i < height; ++i)
    {
        for (int j = 0; j < width; ++j, pixel += desire_channels)
        {
            data[index++] = (unsigned char)(pixel[0]);
            data[index++] = (unsigned char)(pixel[1]);
            data[index++] = (unsigned char)(pixel[2]);
            // Do something with r, g, b
        }
    }

    other(data, out, width, height);

    //Rewrite the image to make sure that image reader is working correctly
    stbi_write_jpg("hola.jpg", width, height, 3, out, width*sizeof(int));
    stbi_image_free(rgb_image);
    free(out);
    free(data);

    return 0;
}