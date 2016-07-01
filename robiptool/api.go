package robiptool

import (
  "os"
  "fmt"
  "io/ioutil"

  "net/http"
)

func FetchBinary(id string) (*os.File, error) {
  url := fmt.Sprintf("http://robip.halake.com/api/%s/latest", id)
  if response, err := http.Get(url); err != nil {
    return nil, err

  } else {
    if body, err := ioutil.ReadAll(response.Body); err != nil {
      return nil, err

    } else {
      if file, err := ioutil.TempFile("", "robip-"); err != nil {
        return nil, err

      } else {
        if _, err := file.Write(body); err != nil {
          return nil, err
        }
        file.Close()

        return file, nil
      }
    }
  }
}
