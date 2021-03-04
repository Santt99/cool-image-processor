#include "images.h"
#include <iostream>
#include <stdint.h>
#include <string.h>

#define STB_IMAGE_IMPLEMENTATION
#include "stb_image.h"

#define STB_IMAGE_WRITE_IMPLEMENTATION
#include "stb_image_write.h"

using namespace std;
int main(int argc, char **argv)
{
    if (argc < 4)
    {
        cout << "Please enter all the required params: filter_type[blur, grayscale] input_image_path output_image_path";
        return 1;
    }

    //Extract parms
    string filter_type = argv[1];
    char *input_image_path = argv[2];
    char *output_image_path = argv[3];

    int width, height, bpp;
    int desire_channels = 3;
    uint8_t *rgb_image =
        stbi_load(input_image_path, &width, &height, &bpp, desire_channels);
    cout << bpp;

    unsigned char *data =
        (unsigned char *)malloc(sizeof(unsigned char) * width * height * 3);
    
        
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

    int outChannels = 3;
    if (filter_type == "grayscale"){
        outChannels = 1;
    }
    
    unsigned char *out = (unsigned char *)malloc(sizeof(unsigned char) * width * height * outChannels);
    filter(data, out, width, height, outChannels);
 
    // Rewrite the image to make sure that image reader is working correctly
    stbi_write_jpg(output_image_path, width, height, outChannels, out, width * sizeof(int));
    stbi_image_free(rgb_image);
    free(out);
    free(data);

    return 0;
}