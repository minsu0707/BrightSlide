# BrightSlide

[![최신 릴리스](https://img.shields.io/github/v/release/minsu0707/BrightSlide?label=%EC%B5%9C%EC%8B%A0%20%EB%A6%B4%EB%A6%AC%EC%8A%A4)](https://github.com/minsu0707/BrightSlide/releases/latest)
![Windows 10/11 x64](https://img.shields.io/badge/Windows-10%20%7C%2011%20(x64)-0078D4)

macOS 메뉴바의 밝기 슬라이더처럼, 트레이 아이콘 클릭 한 번으로 모니터 밝기를 드래그로 조절하는 Windows 전용 유틸리티입니다.

- 트레이 아이콘을 누르면 뜨는 팝업에서 슬라이더를 드래그해 밝기 조절
- 노트북 내장 화면과 외장 모니터를 모두 지원하고, 디스플레이마다 슬라이더가 하나씩 표시
- 아주 가벼움: 실행 파일 약 2MB, 평소 메모리 사용량(작업 집합) 약 1~2MB, 가만히 있을 때 CPU 0%
- 관리자 권한 없이 설치, .NET 같은 런타임 설치도 필요 없음

---

## 다운로드

[**Releases**](https://github.com/minsu0707/BrightSlide/releases/latest)에서 원하는 형태를 받으세요.

| | 설치 프로그램 | 단독 실행 파일 |
|---|---|---|
| 파일 | `BrightSlide-Setup-x.x.x.exe` | `BrightSlide-x.x.x-win-x64.exe` |
| 설치 | 현재 사용자 계정에 설치 (관리자 권한 불필요) | 필요 없음, 더블클릭으로 바로 실행 |
| 시작 메뉴 등록 | O | X |
| 자동 실행 설정 | 설치 중 선택 가능 | 앱 메뉴에서 설정 |
| 제거 | Windows "앱 제거" | 파일 삭제 |
| 업데이트 | 새 버전을 설치하면 덮어씀 | 새 파일로 교체 |

**요구 사항:** Windows 10 또는 11 (x64)

---

## 사용법

1. 실행하면 창 없이 **시스템 트레이**에 아이콘만 표시됩니다.
2. 아이콘을 **클릭**하면 밝기 조절 팝업이 뜹니다.
3. 디스플레이별 슬라이더를 드래그하면 실시간으로 밝기가 바뀝니다.
4. 팝업 밖을 클릭하면 자동으로 닫힙니다.

아이콘이 안 보이면 작업 표시줄의 `^`(숨겨진 아이콘) 안을 확인하세요.

**트레이 아이콘 우클릭 메뉴**

| 메뉴 | 동작 |
|---|---|
| 시작 시 자동 실행 | Windows 로그인 시 BrightSlide를 자동으로 켜도록 설정 |
| 종료 | 프로그램 완전 종료 |

---

## 자동 실행과 재부팅 후 동작

> [!NOTE]
> 자동 실행은 **기본으로 꺼져 있습니다.** 재부팅하면 앱이 자동으로 켜지지 않습니다.

**자동 실행 켜는 방법** (둘 중 하나)

- 설치할 때 "Windows 시작 시 자동 실행"을 체크
- 트레이 아이콘 우클릭, **시작 시 자동 실행** 체크

| 상황 | 동작 |
|---|---|
| 자동 실행을 켠 뒤 | **작업 관리자, 시작 앱** 목록에 BrightSlide가 나타나며, 여기서 사용/사용 안 함으로 바꿀 수도 있음 |
| 자동 실행을 켜기 전 | 시작 앱 목록에 보이지 않는 것이 정상 |
| 설치형을 제거할 때 | 자동 실행 등록도 함께 삭제 |
| 재부팅 후 밝기 값 | 앱은 값을 저장하지 않고 실행할 때 현재 밝기를 읽어 표시함. 유지 여부는 디스플레이(모니터 자체 메모리, Windows 전원 설정)에 따라 다름 |

---

## 지원하는 디스플레이

| 종류 | 제어 방식 | 비고 |
|---|---|---|
| 노트북 내장 디스플레이 | Windows 비디오 밝기 제어(`\\.\LCD`) | 실제 노트북에서는 아직 검증되지 않음 (아래 제한 사항 참고) |
| 외장 모니터 | DDC/CI | 모니터와 연결 케이블(HDMI/DP)이 DDC/CI를 지원해야 동작 |

디스플레이는 실행 시 자동으로 감지되며, 가능한 경우 실제 모니터 모델명을 표시합니다.

---

## 알려진 제한 사항

| 증상 | 원인 / 대처 |
|---|---|
| 노트북 내장 화면이 목록에 나오지 않거나 조절되지 않음 | 내장 디스플레이 지원은 Windows 노트북에서 검증되지 않았음. 문제가 있으면 이슈로 알려주세요 |
| USB-C 허브로 연결한 외장 모니터의 밝기가 조절되지 않음 | 허브를 거치면 DDC/CI 신호가 전달되지 않을 수 있음 |
| 외장 모니터가 내장 화면보다 느리게 따라옴 | DDC/CI 통신 속도(모니터 펌웨어 한계, 명령 1회당 수십 ms) 때문. 빠르게 드래그해도 멈추면 최종 값으로 맞춰짐 |
| 밝기가 매우 낮은 구간(1~8%)에서 값이 건너뜀 | 일부 모니터가 이 구간에서 비선형적으로 동작해서 보정해 회피함 |

---

## 직접 빌드하기

개발자용입니다. Go로 작성되어 있고 외부 라이브러리나 C 컴파일러는 필요 없습니다. **Go 1.27 이상**이 필요합니다.

```powershell
winget install -e --id GoLang.Go --source winget
```

**소스에서 빌드해 실행**

```powershell
git clone https://github.com/minsu0707/BrightSlide.git
cd BrightSlide
$env:CGO_ENABLED = '0'
go build -trimpath -ldflags '-s -w -H=windowsgui' -o BrightSlide.exe .
.\BrightSlide.exe
```

개발할 때 `--show` 옵션을 주면 트레이 클릭 없이 팝업이 바로 열립니다.

**아이콘 리소스 다시 만들기**

`Assets\brightslide.ico`를 바꿨다면 exe에 들어가는 `rsrc_windows_amd64.syso`를 다시 생성하세요.

```powershell
go run ./tools/genrsrc
```

**설치 프로그램 만들기**

[Inno Setup 6](https://jrsoftware.org/isinfo.php)을 설치한 뒤 다음을 실행하세요. 빌드와 패키징이 한 번에 처리되며, 결과물(설치 프로그램과 단독 exe)은 `installer\Output\`에 생성됩니다. 버전은 루트의 `VERSION` 파일에서 읽습니다.

```powershell
winget install -e --id JRSoftware.InnoSetup --source winget
.\installer\build-installer.ps1
```
