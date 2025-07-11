import numpy as np
import matplotlib.pyplot as plt
from scipy.io import wavfile

def read_vad_txt(txt_path):
    """读取 TEN-VAD 输出的 txt 文件，返回每帧是否为语音（0或1）"""
    flags = []
    with open(txt_path, 'r') as f:
        for line in f:
            if line.startswith('[') and ']' in line:
                try:
                    parts = line.strip().split()
                    flag = int(parts[2])
                    flags.append(flag)
                except:
                    continue
    return np.array(flags)

def plot_waveform_and_ten_vad(wav_path, vad_txt_path, frame_size=256, save_path='ten_vad_plot.png'):
    # 读取音频
    sample_rate, audio = wavfile.read(wav_path)
    if audio.ndim > 1:
        audio = audio.mean(axis=1)  # 转为单通道

    audio = audio / np.max(np.abs(audio))  # 归一化
    time_audio = np.arange(len(audio)) / sample_rate

    # 读取 TEN-VAD 结果
    vad_flags = read_vad_txt(vad_txt_path)
    vad_curve = np.repeat(vad_flags, frame_size)
    time_vad = np.arange(len(vad_curve)) / sample_rate

    # 对齐长度
    min_len = min(len(audio), len(vad_curve))
    audio = audio[:min_len]
    vad_curve = vad_curve[:min_len]
    time_audio = time_audio[:min_len]
    time_vad = time_vad[:min_len]

    # 绘图
    plt.figure(figsize=(12, 6))

    # 音频波形
    plt.subplot(2, 1, 1)
    plt.plot(time_audio, audio, linewidth=0.8)
    plt.title("Input audio")
    plt.xlabel("Time in sec")
    plt.ylabel("Normalized Amplitude")

    # TEN-VAD 曲线
    plt.subplot(2, 1, 2)
    plt.plot(time_vad, vad_curve, color='orange', linewidth=0.8)
    plt.title("TEN VAD")
    plt.xlabel("Time in sec")
    plt.ylabel("VAD Result (0 or 1)")
    plt.yticks([0, 1])

    plt.tight_layout()
    plt.savefig(save_path)
    plt.show()

# 示例路径（你可以替换）
wav_path = "../testset/testset-audio-01.wav"
vad_txt_path = "result.txt"

plot_waveform_and_ten_vad(wav_path, vad_txt_path)