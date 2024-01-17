import { userApi } from "../../../../api/api";
import { Job } from "../../../../main.types";
import { Avatar } from "@mui/material";
import JobAcceptButtons from "./JobAcceptButtons";
import { Skeleton } from "@mui/material";

function capitalize(str?: string) {
  return str ? str[0].toUpperCase() + str.slice(1).toLowerCase() : "";
}

export default function DesignerName(props: { designerId: string; job?: Job }) {
  const { data: data } = userApi.useGetUserQuery(props.designerId);
  return (
    <div className=" flex flex-col w-full">
      <div className=" flex flex-row items-center justify-between w-full">
        <div className=" flex flex-row">
          {data ? (
            <Avatar
              className={`outline outline-4 outline-offset-2 !w-24 !h-24 ${
                data.userType == "DESIGNER"
                  ? "outline-designer"
                  : "outline-producer"
              }`}
              alt={`${data.firstName} ${data.lastName}`}
              sx={{ width: 64, height: 64 }}
            >
              {data.firstName.charAt(0)}
            </Avatar>
          ) : (
            <Skeleton variant="circular" width={64} height={64} />
          )}
          {data ? (
            <div className="ml-5 px-4 flex flex-col justify-center">
              <p className=" text-lg">
                {data && data.firstName} {data && data.lastName}
              </p>
              <p className=" text-sm opacity-70">{capitalize(data.userType)}</p>
            </div>
          ) : (
            <div className=" px-4 flex flex-col justify-center">
              <Skeleton variant="rectangular" width={96} height={28} />
              <Skeleton variant="rectangular" width={48} height={20} />
            </div>
          )}
        </div>
        {props.job && <JobAcceptButtons currentJob={props.job} />}
      </div>
    </div>
  );
}
