export const typeName = (cell,_) => {
  switch (cell) {
    case "folder":
      return "フォルダー";
    case "file":
      return "単一ファイル";
    case "http":
      return "Webサーバー";
    case "scp":
      return "SCPサーバー";
    case "sftp":
      return "SFTPサーバー";
  }
  return "";
}

