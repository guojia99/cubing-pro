# https://drive.google.com/file/d/1fm7cZEYs46LzrWroh4Vd6AJFsGDJ9mnq/view?pli=1
import json

csp_pdf = "/home/guojia/worker/code/cube/cubing-pro/build/alg/CSP_v3.1 (1).pdf"


import pdfplumber

# 打开 PDF 文件
csp_data = {

}

def clean_id_text(text):
    cleaned = text.split("#")[0]
    cleaned = cleaned.replace("-", "").strip().lower()
    return cleaned

evStr = "Even (Par)"
oddStr = "Odd (Impar)"

with pdfplumber.open(csp_pdf) as pdf:
    last = ""
    start = False

    lastEv = ""
    lastOdd = ""

    def updateLast():
        if last != "":
            lev =  [i for i in lastEv.split(" ") if i != ""]
            lod = [i for i in lastOdd.split(" ") if i != ""]
            # 可能存在无对称型


            dd = {
                "base": {
                    "even": lev[0],
                    "odd": lod[0],
                },
            }
            if len(lev) == 2 and len(lod) == 2:
                dd["mirror"] = {
                    "even": lev[1],
                    "odd": lod[1],
                }


            csp_data[last] = dd



    for page in pdf.pages:
        st = page.extract_text().split("\n")
        for data in st:
            if "CASES" in data:
                start = True

            if not start:
                continue

            if evStr in data:
                lastEv = data.replace(evStr, "")
            if oddStr in data:
                lastOdd = data.replace(oddStr, "")

            if "#" in data:
                updateLast()
                last = clean_id_text(data)
    updateLast()

print(json.dumps(csp_data, indent=4))