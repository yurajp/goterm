package goterm

import (
  "os"
  "os/exec"
  "fmt"
  "encoding/json"
  "errors"
  "strings"
  "regexp"
  "net/url"
  "io/ioutil"
  "time"
)

var recPath = "/storage/emulated/0/recording"

type Auth struct {
  Errors []string `json:"errors"`
  Failed int `json:"failed_attempts"`
  Result string `json:"auth_result"`
}

type Dialog struct {
  Code int `json:"code"`
  Text string `json:"text"`
}

type DialogT struct {
  Code int `json:"code"`
  Text string `json:"text"`
  Index int `json:"index"`
}

type Value struct {
  Index int `json:"index"`
  Text string `json:"text"`
}

type DialogM struct {
  Code int `json:"code"`
  Text string `json:"text"`
  Values []Value `json:"values"`
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

func check(e error) bool {
  if e != nil {
    fmt.Println(e)
    return true
  }
  return false
}

func (p Place) String() string {
  la, err := p.Field(1)
  if check(err) {
    return ""
  }
  lo, err := p.Field(2)
  if check(err) {
    return ""
  }
  sp, err := p.Field(3)
  if check(err) {
    return ""
  }
  return fmt.Sprintf("latitude: %s, longitude: %s, speed:  %s km/h", la, lo, sp)
}

func (p Place) Field(i int) (string, error) {
  if i == 0 || i > 3 {
    return "", errors.New("Args should be 1 || 2 || 3")
  }
  if i == 1 {
    return fmt.Sprintf("%f", p.Latitude), nil
  }
  if i == 2 {
    return fmt.Sprintf("%f", p.Longitude), nil
  }
  if i == 3 {
    return fmt.Sprintf("%s", MtoKm(p.Speed)), nil
  }
  return "", nil
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

//Displays error in floating window.
func ToastErr(err error) {
  s := fmt.Sprintf("%s", err)
  exec.Command("termux-toast", "-b", "red", s).Run()
}

func Notification(txt string) {
  exec.Command("termux-notification", "--id", "termapp", "--ongoing", "-c", txt).Run()
}

func Vibrate() {
  exec.Command("termux-vibrate").Run()
}

func Confirm() bool {
  cmd := exec.Command("termux-confirm")
  var dlg Dialog
  out, _ := cmd.Output()
  json.Unmarshal(out, &dlg)
  return dlg.Text == "yes"
}

func TimeName() string {
  return time.Now().Format("060102-150405")
}

func Photo(p string) error {
  cmd := exec.Command("termux-camera-photo", p)
  err := cmd.Run()
  if err != nil {
    return err
  }
  return nil
}

//Returns user short text
func UserName() (string, error) {
  cmd := exec.Command("termux-dialog", "-t", "  NAME")
  out, err := cmd.Output()
  if err != nil {
    return "", err
  }
  var dlg Dialog
  err = json.Unmarshal(out, &dlg)
  if err != nil {
    return "", err
  }
  return dlg.Text, nil
}

//Returns user text
func UserText() (string, error) {
  cmd := exec.Command("termux-dialog", "-m", "-t", "  TEXT")
  out, err := cmd.Output()
  if err != nil {
    return "", err
  }
  var dlg DialogT
  err = json.Unmarshal(out, &dlg)
  if err != nil {
    return "", err
  }
  if dlg.Code != -1 {
    return "", errors.New("Canceled")
  }
  return dlg.Text, nil
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
  if len(list) == 0 {
    return []Contact{}, errors.New(" EMPTY CONTACTS")
  }
  return list, nil
}

func GetContact(name string) (Contact, error) {
  if name == "" {
    return Contact{}, errors.New("No Name")
  }
  cnts, err := ContactList()
  if err != nil {
    return Contact{}, err
  }
  matches := []Contact{}
  for _, cnt := range cnts {
    if strings.Contains( strings.ToLower(cnt.Name), strings.ToLower(name)) {
      matches = append(matches, cnt)
    }
  }
  if len(matches) == 0 {
    return Contact{}, errors.New("Contact not found")
  }
  if len(matches) == 1 {
    return matches[0], nil
  }
  names := []string{}
  for _, c := range matches {
    names = append(names, c.Name)
  }
  _, idx, err := Radio(names)
  if err != nil {
    return Contact{}, err
  }
  return matches[idx], nil
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
    return "", errors.New("No sms")
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

func SendSms(cnt Contact, txt string) error {
  cmd := exec.Command("termux-sms-send", "-n", cnt.Number, txt)
  err := cmd.Run()
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
  dl := "%2C"
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

func NavigatorLink() (string, error) {
  p, err := GetLocation()
  if err != nil {
    return "", err
  }
  la, err := p.Field(1)
  if check(err) {
    return "", err
  }
  lo, err := p.Field(2)
  if check(err) {
    return "", err
  }
  dl := ","
  params := fmt.Sprintf("%s%s%s", la, dl, lo)
  return "https://maps.yandex.ru/?text=" + params, nil
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
  cmd := exec.Command("termux-tts-speak", "-r", "0.9", "-p", "0.9", text)
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

//Function calls to contact.
func Call(cnt Contact) error {
  cmd := exec.Command("termux-telephony-call", cnt.Number)
  err := cmd.Run()
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

func WaSend(cnt Contact, message string) error {
  phone := cnt.Number
  if strings.HasPrefix(phone, "8") {
    phone = strings.Replace(phone, "8", "7", 1)
  }
  text := url.QueryEscape(message)
  lnk := fmt.Sprintf("https://api.whatsapp.com/send?phone=%s&text=%s", phone, text)
  return OpenUrl(lnk)
}
  
func HandleRecord(recDir string) error {
  lst, err := ioutil.ReadDir(recDir)
  if err != nil {
    return err
  }
  i := len(lst)
  var newname string
  for {
    num := fmt.Sprintf("%02d", i +1)
    newname = recDir + "/rec-" + num + ".m4a"
    if _, err = os.Stat(newname); err != nil {
      if os.IsNotExist(err) {
        break
      }
    }
    i++
  }
  err = os.Rename(recPath, newname)
  if err != nil {
    return fmt.Errorf("Moving: %s", err)
  }
  return nil
}

func Record(dir string) error {
  cmd := exec.Command("termux-microphone-record", "-d", "-f", recPath)
  err := cmd.Run()
  if err != nil {
    return err
  }
  fmt.Println("\trecording...")
  fmt.Println("\tEnter to stop")
  var q string
  fmt.Scanf("%s", &q)
  exec.Command("termux-microphone-record", "-q").Run()
  return HandleRecord(dir)
}

func Play(p string) error {
  cmd := exec.Command("termux-media-player", "play", p)
  err := cmd.Run()
  if err != nil {
    return err
  }
  trp := strings.Split(p, "/")
  tr := trp[len(trp) - 1]
  fmt.Println("\tplaying ", tr)
  fmt.Println("\tEnter to stop")
  var s string
  fmt.Scanf("%s", &s)
  return exec.Command("termux-media-player", "stop").Run()
}

func Radio(sl []string) (string, int, error) {
  t := "  SELECT"
  v := strings.Join(sl, ",")
  cmd := exec.Command("termux-dialog", "radio", "-t", t, "-v", v)
  out, err := cmd.Output()
  if err!= nil {
    return "", 0, err
  }
  var rad DialogT
  err = json.Unmarshal(out, &rad)
  if err != nil {
    return "", 0, err
  }
  if rad.Code != -1 {
    return "", 0, errors.New("Cannot select")
  }
  return rad.Text, rad.Index, nil
}

func Checkbox(sl []string) ([]string, []int, error) {
  t := "  SELECT"
  v := strings.Join(sl, ",")
  cmd := exec.Command("termux-dialog", "checkbox", "-t", t, "-v", v)
  out, err := cmd.Output()
  if err!= nil {
    return []string{}, []int{}, err
  }
  var cbx DialogM
  err = json.Unmarshal(out, &cbx)
  if err != nil {
    return []string{}, []int{}, err
  }
  if cbx.Code != -1 {
    return []string{}, []int{}, errors.New("Cannot select")
  }
  vals := []string{}
  idxs := []int{}
  for _, v := range cbx.Values {
    vals = append(vals, v.Text)
    idxs = append(idxs, v.Index)
  }
  return vals, idxs, nil
}
