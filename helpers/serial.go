package helpers

import (
    "time"
    "errors"

    "blaulicht/config"

    "github.com/op/go-logging"
    "github.com/tarm/serial"
)

var log = logging.MustGetLogger("blaulicht")
var Conf config.Config

type SerialPort struct {
  *serial.Port
}

func NewSerialPort() (*SerialPort, error) {
  serialPort, err := serial.OpenPort(&serial.Config{Name: Conf.SerialPortDevice, Baud: Conf.SerialPortBaudRate, ReadTimeout: time.Second * 5})
  if err != nil {
    log.Errorf("Unable to open serial port: %v", err.Error())
    return &SerialPort{}, err
  }
  
  return &SerialPort{serialPort}, nil
}

func (s *SerialPort) SendCommand(command string) error {
  //send command
  _, err := s.Write([]byte(command + "\r\n"))
  if err != nil {
    log.Errorf("Error sending command: %s", err.Error())
    return err
  }
  log.Debugf("Sent command: %s", command)

  //read response
  buf := make([]byte, 128)
  n, err := s.Read(buf)
  if err != nil {
    log.Errorf("Unable to read from serial line: %s", err.Error())
    return err
  }
  log.Debugf("Serial response: %s", string(buf[:n])) 
  if string(buf[:n]) != command + "\r\n" {
    log.Errorf("Command didn't succeed")
    return errors.New("Arduino returned an error")
  }
  
  return nil
} 