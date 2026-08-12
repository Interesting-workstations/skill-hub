/** 通用弹窗：受控组件，含确认/取消按钮；danger 表示危险操作。 */

import type { ReactNode } from "react";

interface Props {
  open: boolean;
  title: string;
  onClose: () => void;
  onConfirm?: () => void;
  confirmText?: string;
  cancelText?: string;
  danger?: boolean;
  size?: "default" | "lg";
  children?: ReactNode;
}

export default function AppDialog({
  open,
  title,
  onClose,
  onConfirm,
  confirmText = "确定",
  cancelText = "取消",
  danger,
  size = "default",
  children,
}: Props) {
  if (!open) return null;
  return (
    <div className="modal-mask" onClick={onClose}>
      <div className={`modal ${size === "lg" ? "modal-lg" : ""}`} onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>{title}</h3>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>
        <div className="modal-body">{children}</div>
        {(onConfirm || danger) && (
          <div className="modal-footer">
            <button className="btn" onClick={onClose}>{cancelText}</button>
            <button className={`btn ${danger ? "btn-danger-solid" : "btn-primary"}`} onClick={onConfirm}>
              {confirmText}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
