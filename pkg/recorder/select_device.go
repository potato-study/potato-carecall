package recorder

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
)

var selectedDeviceID = -1

// 입력 장치 선택
func InitRecordDevice() error {
	if selectedDeviceID >= 0 {
		return nil
	}
	// 입력 장치 목록 가져오기
	devices, err := portaudio.Devices()
	if err != nil {
		return fmt.Errorf("입력 장치 목록 가져오기 중 오류 발생: %v", err)
	}

	// 입력 장치 목록 출력
	fmt.Println("사용 가능한 입력 장치 목록:")
	for i, device := range devices {
		if device.MaxInputChannels > 0 {
			fmt.Printf("%d: %s (기본 샘플 레이트: %f)\n", i, device.Name, device.DefaultSampleRate)
		}
	}

	// 사용자로부터 장치 선택 받기
	var deviceIndex int
	fmt.Print("사용할 입력 장치 번호를 입력하세요: ")
	_, err = fmt.Scanf("%d", &deviceIndex)
	if err != nil || deviceIndex < 0 || deviceIndex >= len(devices) || devices[deviceIndex].MaxInputChannels == 0 {
		return fmt.Errorf("잘못된 장치 번호입니다")
	}
	selectedDeviceID = deviceIndex

	return nil
}

func getDevice() (*portaudio.DeviceInfo, error) {
	if selectedDeviceID < 0 {
		return nil, fmt.Errorf("장치 ID가 설정되지 않았습니다")
	}

	devices, err := portaudio.Devices()
	if err != nil {
		return nil, fmt.Errorf("입력 장치 목록 가져오기 중 오류 발생: %v", err)
	}
	return devices[selectedDeviceID], nil
}
