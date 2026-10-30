# BrightSlide

macOS 메뉴바의 밝기 슬라이더처럼, 트레이 아이콘 클릭 한 번으로 모니터 밝기를 드래그로 조절하는 Windows 전용 유틸리티입니다.

## 다운로드

[Releases](https://github.com/minsu0707/BrightSlide/releases)에서 원하는 형태를 받으세요 (.NET 설치 불필요).

- **설치 프로그램** `BrightSlide-Setup-x.x.x.exe`: 관리자 권한 없이 현재 사용자 계정에 설치되며, 시작 메뉴 등록과 선택적 자동 실행 설정을 지원하고 Windows의 "앱 제거"에서 삭제할 수 있습니다. 새 버전 설치 시 기존 버전 위에 덮어씁니다.
- **단독 실행 파일** `BrightSlide-x.x.x-win-x64.exe`: 설치 없이 더블클릭하면 바로 동작합니다.

## 사용법

1. 실행하면 창 없이 시스템 트레이에 아이콘만 표시됩니다.
2. 트레이 아이콘을 클릭하면 밝기 조절 팝업이 뜹니다.
3. 연결된 디스플레이마다 슬라이더가 하나씩 보이며, 드래그하면 실시간으로 밝기가 바뀝니다.
4. 팝업 밖을 클릭하면 자동으로 닫힙니다.
5. 트레이 아이콘을 **우클릭**하면 다음 메뉴를 사용할 수 있습니다.
   - **시작 시 자동 실행**: Windows 로그인 시 BrightSlide가 자동으로 켜지도록 설정
   - **종료**: 프로그램 완전 종료

## 자동 실행과 재부팅 후 동작

- **자동 실행은 기본으로 꺼져 있습니다.** 재부팅하면 앱이 자동으로 켜지지 않으며, 시작 메뉴에서 직접 실행하거나 아래 방법으로 자동 실행을 켜야 합니다.
  - 설치할 때 "Windows 시작 시 자동 실행"을 체크하거나
  - 트레이 아이콘 우클릭 → **시작 시 자동 실행**을 체크
- 자동 실행을 켜면 **작업 관리자 → 시작 앱** 목록에 BrightSlide가 나타나고, 여기서 사용/사용 안 함으로 바꿀 수도 있습니다. 켜기 전에는 목록에 보이지 않는 것이 정상입니다.
- 앱은 밝기 값을 따로 저장하지 않습니다. 실행할 때 디스플레이의 현재 밝기를 읽어 슬라이더에 표시하며, 재부팅 후에도 값이 유지되는지는 디스플레이(모니터 자체 메모리, Windows 전원 설정)에 따라 달라집니다.
- 설치형을 제거하면 자동 실행 등록도 함께 삭제됩니다.

## 지원하는 디스플레이

| 종류 | 제어 방식 | 비고 |
|---|---|---|
| 노트북 내장 디스플레이 | Windows 전원 관리(WMI) | 대부분의 노트북에서 동작 |
| 외장 모니터 | DDC/CI | 모니터와 연결 케이블(HDMI/DP)이 DDC/CI를 지원해야 동작 |

디스플레이는 실행 시 자동으로 감지되며, 가능한 경우 실제 모니터 모델명을 표시합니다.

## 요구 사항

- Windows 10 또는 11 (x64)

## 알려진 제한 사항

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

설치 프로그램을 만들려면 [Inno Setup 6](https://jrsoftware.org/isinfo.php)을 설치한 뒤 다음을 실행하세요. 빌드와 패키징이 한 번에 처리되며 결과물은 `installer\Output\`에 생성됩니다.

```powershell
winget install -e --id JRSoftware.InnoSetup --source winget
.\installer\build-installer.ps1
```
