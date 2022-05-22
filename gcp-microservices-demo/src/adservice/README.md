# Adservice

The Adservice provides advertisement based on context keys. If no context keys are provided then it returns random ads.

use this command to zip fission Java function: 
```
zip adservice.zip -r src/ pom.xml
``` 
**Do not include other files or folders when zipping fission Java function!**

**NOTE: Do not use fastjson, use jackson!** Because "spring-boot-starter-web" package comes with Jackson's core class library("spring-boot-starter-web" package is used by default for writing fission Java function), and dependency conflicts will occur when importing fastjson(will throw an error: java.lang.NoClassDefFoundError: retrofit2/Converter$Factory).  