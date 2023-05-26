package goterm

import (
  "os/exec"
  "fmt"
  "encoding/json"
  "errors"
  "strings"
  "regexp"
)

type Auth struct {
  Errors []string `json:"errors"`
  Failed int `json:"failed_attempts"`
  Result string `json:"auth_result"`
}

type Location struct {
  Latitude float64 `json:"latitude"`
  Longitude float64 `json:"longitude"`
  Altitude float64 `json:"altitude"`
  Accuracy float64 `json:"accuracy"`
  Vertacc float64 `json:"vertical_accuracy"`
  Bearing float64 `json:"bearing"`
  Speed float64 `json:"speed"`
  ElapsedMs int `json:"elapsedMs"`
  Provider string `json:"provider"`
}

type Place struct {
  Latitude float64
  Longitude float64
  Speed float64
}

type Sms struct {
  Idth int `json:"threadid"`
  Type string `json:"type"`
  Read bool `json:"read"`
  Number string `json:"number"`
  Time string `json:"received"`
  Body string `json:"body"`
  Id int `json:"_id"`
}

type Contact struct {
   Name string `json:"name"`
   Number string `json:"number"`
}

//Returns list of contacts on the phone.1
func ContactList() ([]Contact, error) {
  cmd := exec.Command("termux-contact-list")
  jsn, err := cmd.Output()
  if err != nil {
    return []Contact{}, err
  }
  var list []Contact
  err = json.Unmarshal(jsn, &list)
  if err != nil {
    return []Contact{}, err
  }
  return list, nil
}

//Returns phone number of contact or returns error.
func GetNumber(name string) (string, error) {
  cnts, err := ContactList()
  if err != nil {
    return "", err
  }
  for _, cnt := range cnts {
    if strings.Contains( strings.ToLower(cnt.Name), strings.ToLower(name)) {
      return cnt.Number, nil
    }
  }
  return "", errors.New("Contact not found")
}

//Returns text of last sms.
func LastSms() (string, error) {
  cmd := exec.Command("termux-sms-list", "-l", "1")
  jsn, err := cmd.Output()
  if err != nil {
    return "", err
  }
  var list []Sms
  err = json.Unmarshal(jsn, &list)
  if err != nil {
    return "", err
  }
  if len(list) == 0 {
    return "", errors.New("Empty contacts")
  }
  return list[0].Body, nil
}

//Just copies string to clipboard.
func Copy(s string) {
  cmd := exec.Command("termux-clipboard-set", s)
  cmd.Run()
}

//Gets digital code from string
func GetCode(sms string) string {
  re := regexp.MustCompile(`\d{4,7}\.?`)
  words := strings.Fields(sms)
  for _, w := range words {
    if re.MatchString(w) {
      return strings.TrimRight(w, ".")
    }
  }
  return ""
}

//Copies digital code from last sms to clipboard displaying sms in floating window.
func CopySms() error {
  sms, err := LastSms()
  if err != nil {
    return err
  }
  code := GetCode(sms)
  if code == "" {
    Copy(sms)
  } else {
    Copy(code)
  }
  err = exec.Command("termux-toast", "-b", "black", sms).Run()
  if err != nil {
    return err
  }
  return nil
}

//Function returns coordinates or returns error
func GetLocation() (Place, error) {
  cmd := exec.Command("termux-location", "-p", "network")
  jsn, err := cmd.Output()
  if err != nil {
    return Place{}, err
  }
  var lc Location
  err = json.Unmarshal(jsn, &lc)
  if err != nil {
    return Place{}, err
  }
  p := Place{lc.Latitude, lc.Longitude, lc.Speed}
  return p, nil
}  

//Opens geoposition in Google Maps or returns error. Prints not zero speed to console.
func Locate() error {
  p, err := GetLocation()
  if err != nil {
    return fmt.Errorf("GetLocation: %v\n", err)
  }
  la := p.Latitude
  lo := p.Longitude
  sp := p.Speed
  dl := "%20"
  url := fmt.Sprintf("https://google.com/maps?q=%f%s%f", la, dl, lo)
  err = OpenUrl(url)
  if err != nil {
    return fmt.Errorf("GoogleMaps: %v\n", err)
  }
  if sp != 0 {
    fmt.Printf("SPEED %s km/h", MtoKm(sp))
  }
  return nil
}

//Transcribes m/sec to km/hour.
func MtoKm(s float64) string {
  kmh := s * 3.6
  return fmt.Sprintf("%d", int(kmh))
}

//Function check fingerprint. If authorization is success returns true else returns error.
func Fingerprint() (bool, error) {
  cmd := exec.Command("termux-fingerprint")
  jsn, err := cmd.Output()
  if err != nil {
    return false, err
  }
  var auth Auth
  err = json.Unmarshal(jsn, &auth)
  if err != nil {
    return false, err
  }
  ok := "AUTH_RESULT_SUCCESS"
  if len(auth.Errors) == 0 && auth.Result == ok {
    return true, nil
  }
  fmt.Println(auth.Result)
  return false, errors.New("Fingerprint Error!")
}

//Speak the text of string. If unsuccess returns error.
func Speak(text string) error {
  if len(text) == 0 {
    return errors.New("Nothing to speak")
  }
  cmd := exec.Command("termux-tts-speak", text)
  err := cmd.Run()
  if err != nil {
    return err
  }
  return nil
}

//Opens url in brawser. If unsuccess returns error.
func OpenUrl(url string) error {
  cmd := exec.Command("termux-open-url", url)
  err := cmd.Run()
  if err != nil {
    return err
  }
  return nil
}

//Function searhes phone number for contact name and makes call. If unsuccess returns error.
func Call(name string) error {
  number, err := GetNumber(name)
  if err != nil {
    return err
  }
  cmd := exec.Command("termux-telephony-call", number)
  err = cmd.Run()
  if err != nil {
    return err
  }
  return nil
}

//Function translates speech to string. If result is empty returns error.
func Speech() (string, error) {
  cmd := exec.Command("termux-speech-to-text")
  text, err := cmd.Output()
  if err != nil {
    return "", err
  }
  if string(text) == "\n" || string(text) == "" {
    return "", errors.New(" Not responding")
  }
  return string(text), nil
}