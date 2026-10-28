# BrightSlide

macOS 메뉴바의 밝기 슬라이더처럼, 트레이 아이콘 클릭 한 번으로 모니터 밝기를 드래그로 조절하는 Windows 전용 유틸리티입니다.

## 다운로드

[Releases](https://github.com/minsu0707/BrightSlide/releases)에서 최신 버전의 `BrightSlide-x.x.x-win-x64.exe`를 받아 바로 실행하세요. 설치 과정 없이 더블클릭하면 바로 동작합니다 (.NET 설치 불필요).

## 사용법

1. 실행하면 창 없이 시스템 트레이에 아이콘만 표시됩니다.
2. 트레이 아이콘을 클릭하면 밝기 조절 팝업이 뜹니다.
3. 연결된 디스플레이마다 슬라이더가 하나씩 보이며, 드래그하면 실시간으로 밝기가 바뀝니다.
4. 팝업 밖을 클릭하면 자동으로 닫힙니다.
5. 트레이 아이콘을 **우클릭**하면 다음 메뉴를 사용할 수 있습니다.
   - **시작 시 자동 실행**: Windows 로그인 시 BrightSlide가 자동으로 켜지도록 설정
   - **종료**: 프로그램 완전 종료

## 지원하는 디스플레이

| 종류 | 제어 방식 | 비고 |
|---|---|---|
| 노트북 내장 디스플레이 | Windows 전원 관리(WMI) | 대부분의 노트북에서 동작 |
| 외장 모니터 | DDC/CI | 모니터와 연결 케이블(HDMI/DP)이 DDC/CI를 지원해야 동작 |

디스플레이는 실행 시 자동으로 감지되며, 가능한 경우 실제 모니터 모델명을 표시합니다.

## 요구 사항

- Windows 10 또는 11 (x64)

## 알려진 제한 사항

- 설치형 인스톨러는 아직 없고, 실행 파일(.exe) 형태로만 배포됩니다.
- USB-C 허브를 경유해 연결된 외장 모니터는 DDC/CI 신호가 전달되지 않아 밝기 조절이 안 될 수 있습니다.
- 외장 모니터는 DDC/CI 통신 속도(모니터 펌웨어 한계, 보통 명령 1회당 수십 ms) 때문에 내장 디스플레이보다 밝기 변경이 느리게 느껴질 수 있습니다.
- 일부 모니터는 밝기 값이 매우 낮은 구간에서 비선형적으로 동작할 수 있어, 해당 구간은 보정해서 회피하고 있습니다.

## 직접 빌드하기

개발자용 — 소스에서 직접 빌드하려면:

```powershell
git clone https://github.com/minsu0707/BrightSlide.git
cd BrightSlide
dotnet run
```

단일 실행 파일로 배포판을 만들려면:

```powershell
dotnet publish -c Release -r win-x64 --self-contained true -p:PublishSingleFile=true -p:IncludeNativeLibrariesForSelfExtract=true
```

.NET 9 SDK가 필요합니다.
